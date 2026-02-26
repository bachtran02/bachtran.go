package libs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"

	"github.com/bachtran02/bachtran.go/models"
	md "github.com/bachtran02/bachtran.go/models"
)

type PrometheusClient struct {
	nodesConfig []md.NodeConfig
}

func NewPrometheusClient(nodesConfig []md.NodeConfig) *PrometheusClient {
	return &PrometheusClient{
		nodesConfig: nodesConfig,
	}
}

func (pc *PrometheusClient) FetchNodesStatus(ctx context.Context) ([]models.NodeStatus, error) {
	if len(pc.nodesConfig) == 0 {
		return nil, fmt.Errorf("no nodes configured")
	}

	snap1 := pc.scrapeAllNodes(ctx)
	time.Sleep(1 * time.Second)
	snap2 := pc.scrapeAllNodes(ctx)

	var finalStatuses []models.NodeStatus
	for _, node := range pc.nodesConfig {
		s1, ok1 := snap1[node.Name]
		s2, ok2 := snap2[node.Name]

		var status models.NodeStatus
		if !ok1 || !ok2 {
			status = models.NodeStatus{Online: false, SystemInfo: models.SystemInfo{Hostname: node.Name}}
		} else {
			status = pc.buildNodeStatusWithDelta(s1, s2)
			status.SystemInfo.Hostname = node.Name
			status.Online = true
		}
		finalStatuses = append(finalStatuses, status)
	}
	return finalStatuses, nil
}

func (pc *PrometheusClient) scrapeAllNodes(ctx context.Context) map[string]map[string]float64 {
	var wg sync.WaitGroup
	mu := sync.Mutex{}
	results := make(map[string]map[string]float64)

	for _, node := range pc.nodesConfig {
		wg.Add(1)
		go func(n md.NodeConfig) {
			defer wg.Done()

			req, _ := http.NewRequestWithContext(ctx, "GET", n.NodeExporterUrl, nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			metrics, err := parsePrometheusMetrics(string(body))
			metrics["timestamp"] = float64(time.Now().Unix())
			if err != nil {
				return
			}
			mu.Lock()
			results[n.Name] = metrics
			mu.Unlock()
		}(node)
	}
	wg.Wait()
	return results
}

func parsePrometheusMetrics(data string) (map[string]float64, error) {
	metrics := make(map[string]float64)

	parser := expfmt.NewTextParser(model.UTF8Validation)
	metricFamilies, err := parser.TextToMetricFamilies(strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	for name, mf := range metricFamilies {
		for _, m := range mf.GetMetric() {
			var value float64

			// Extract value based on type
			if m.Gauge != nil {
				value = m.GetGauge().GetValue()
			} else if m.Counter != nil {
				value = m.GetCounter().GetValue()
			} else if m.Untyped != nil {
				value = m.GetUntyped().GetValue()
			} else {
				// Skipping summaries and histograms for a simple map
				continue
			}
			// Build metric key with labels
			labelStr := ""
			if labels := m.GetLabel(); len(labels) > 0 {
				var labelPairs []string
				for _, l := range labels {
					labelPairs = append(labelPairs, fmt.Sprintf("%s=\"%s\"", l.GetName(), l.GetValue()))
				}
				labelStr = "{" + strings.Join(labelPairs, ",") + "}"
			}
			metrics[name+labelStr] = value
		}
	}
	return metrics, nil
}

func (pc *PrometheusClient) buildNodeStatusWithDelta(m1, m2 map[string]float64) models.NodeStatus {
	status := buildNodeStatus(m2)

	var (
		totalDelta, idleDelta float64
		rxDelta, txDelta      float64
		timeDelta             = m2["timestamp"] - m1["timestamp"]
	)

	for key, val2 := range m2 {
		/* --- CPU logic --- */
		if strings.HasPrefix(key, "node_cpu_seconds_total") {
			if val1, ok := m1[key]; ok {
				diff := val2 - val1
				totalDelta += diff
				if strings.Contains(key, "mode=\"idle\"") {
					idleDelta += diff
				}
			}
		}
		/* --- Network logic --- */
		if !strings.Contains(key, "device=\"lo\"") {
			if strings.HasPrefix(key, "node_network_receive_bytes_total") {
				if val1, ok := m1[key]; ok {
					rxDelta += val2 - val1
				}
			}
			if strings.HasPrefix(key, "node_network_transmit_bytes_total") {
				if val1, ok := m1[key]; ok {
					txDelta += val2 - val1
				}
			}
		}
	}
	if totalDelta > 0 {
		status.CPU.UsagePercent = (1.0 - (idleDelta / totalDelta)) * 100
	}
	if timeDelta > 0 {
		status.Network.DownSpeed = (rxDelta / timeDelta * 8) / 1_000_000
		status.Network.UpSpeed = (txDelta / timeDelta * 8) / 1_000_000
	}
	return status
}

func buildNodeStatus(metrics map[string]float64) models.NodeStatus {
	status := models.NodeStatus{}

	// CPU metrics
	if _, ok := metrics[`node_cpu_seconds_total{cpu="0",mode="idle"}`]; ok {
		var (
			cores     = 0
			maxFreqHz float64
		)
		for key, val := range metrics {
			if strings.HasPrefix(key, "node_cpu_seconds_total") && strings.Contains(key, `mode="idle"`) {
				cores++
			}
			if strings.HasPrefix(key, "node_cpu_frequency_max_hertz") && strings.Contains(key, `cpu="0"`) {
				maxFreqHz = val
			}
		}
		status.SystemInfo.Cores = cores
		status.SystemInfo.MaxFreqGHz = maxFreqHz / 1e9

		// Calculate CPU usage
		var totalIdle, totalBusy float64
		for key, value := range metrics {
			if strings.HasPrefix(key, "node_cpu_seconds_total") {
				if strings.Contains(key, "mode=idle") {
					totalIdle += value
				} else {
					totalBusy += value
				}
			}
		}
		if totalIdle+totalBusy > 0 {
			status.CPU.UsagePercent = (totalBusy / (totalIdle + totalBusy)) * 100
		}
	}

	// Temperature (if available)
	for key, value := range metrics {
		if strings.Contains(key, "node_hwmon_temp_celsius") && strings.Contains(key, "chip=") {
			status.CPU.Temperature = value
			break
		}
	}

	// Memory metrics
	if memTotal, ok := metrics["node_memory_MemTotal_bytes"]; ok {
		status.Memory.Total = uint64(memTotal)
		status.SystemInfo.TotalRAM = uint64(memTotal)
	}
	if memAvail, ok := metrics["node_memory_MemAvailable_bytes"]; ok {
		status.Memory.Available = uint64(memAvail)
		status.Memory.Used = status.Memory.Total - status.Memory.Available
		if status.Memory.Total > 0 {
			status.Memory.UsedPercent = (float64(status.Memory.Used) / float64(status.Memory.Total)) * 100
		}
	}

	// Disk metrics (root filesystem)
	for key, value := range metrics {
		if strings.Contains(key, "node_filesystem_size_bytes") && (strings.Contains(key, `mountpoint="/"`) || strings.Contains(key, "fstype=ext4")) {
			status.Disk.Total = uint64(value)
			status.SystemInfo.TotalDisk = uint64(value)
			// Find corresponding available bytes
			availKey := strings.Replace(key, "size_bytes", "avail_bytes", 1)
			if avail, ok := metrics[availKey]; ok {
				status.Disk.Available = uint64(avail)
				status.Disk.Used = status.Disk.Total - status.Disk.Available
				if status.Disk.Total > 0 {
					status.Disk.UsedPercent = (float64(status.Disk.Used) / float64(status.Disk.Total)) * 100
				}
			}
			break
		}
	}

	// System uptime
	if bootTime, ok := metrics["node_boot_time_seconds"]; ok {
		uptime := time.Since(time.Unix(int64(bootTime), 0)).Seconds()
		status.Uptime = uptime
	}

	for key := range metrics {
		if strings.HasPrefix(key, "node_exporter_build_info") {
			if strings.Contains(key, "os=\"") {
				parts := strings.Split(key, "os=\"")
				status.SystemInfo.OS = strings.Split(parts[1], "\"")[0]
			}
			if strings.Contains(key, "arch=\"") {
				parts := strings.Split(key, "arch=\"")
				status.SystemInfo.Architecture = strings.Split(parts[1], "\"")[0]
			}
		}
	}
	return status
}
