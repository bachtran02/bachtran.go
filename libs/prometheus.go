package libs

type PrometheusClient struct {
	nodesConfig []NodeConfig
}

func NewPrometheusClient(nodesConfig []NodeConfig) *PrometheusClient {
	return &PrometheusClient{
		nodesConfig: nodesConfig,
	}
}

// func (pc *PrometheusClient) FetchNodesStatus() ([]models.NodeStatus, error) {
// 	var (
// 		statuses []models.NodeStatus
// 	)

// 	if len(pc.nodesConfig) == 0 {
// 		return nil, fmt.Errorf("no nodes configured for Prometheus client")
// 	}

// 	for _, node := range pc.nodesConfig {
// 		resp, err := http.Get(node.NodeExporterUrl + "/metrics")
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to fetch metrics for node %s: %w", node.Name, err)
// 		}
// 		defer resp.Body.Close()

// 		body, err := io.ReadAll(resp.Body)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to read response body for node %s: %w", node.Name, err)
// 		}

// 		metrics, err := parsePrometheusMetrics(string(body))
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to parse metrics for node %s: %w", node.Name, err)
// 		}

// 		statuses = append(statuses, buildHomelabStatus(metrics))
// 	}

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read response body: %w", err)
// 	}

// 	metrics, err := parsePrometheusMetrics(string(body))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse metrics: %w", err)
// 	}

// 	return buildHomelabStatus(metrics), nil
// }

// func parsePrometheusMetrics(data string) (map[string]float64, error) {
// 	metrics := make(map[string]float64)

// 	parser := expfmt.NewTextParser(model.UTF8Validation)
// 	metricFamilies, err := parser.TextToMetricFamilies(strings.NewReader(data))
// 	if err != nil {
// 		return nil, err
// 	}

// 	for name, mf := range metricFamilies {
// 		// mf.GetMetric() returns []*dto.Metric
// 		for _, m := range mf.GetMetric() {
// 			var value float64

// 			// Extract value based on type
// 			if m.Gauge != nil {
// 				value = m.GetGauge().GetValue()
// 			} else if m.Counter != nil {
// 				value = m.GetCounter().GetValue()
// 			} else if m.Untyped != nil {
// 				value = m.GetUntyped().GetValue()
// 			} else {
// 				// Skipping summaries and histograms for a simple map
// 				continue
// 			}

// 			// Build metric key with labels
// 			labelStr := ""
// 			if labels := m.GetLabel(); len(labels) > 0 {
// 				var labelPairs []string
// 				for _, l := range labels {
// 					labelPairs = append(labelPairs, fmt.Sprintf("%s=\"%s\"", l.GetName(), l.GetValue()))
// 				}
// 				labelStr = "{" + strings.Join(labelPairs, ",") + "}"
// 			}

// 			metrics[name+labelStr] = value
// 		}
// 	}

// 	return metrics, nil
// }

// func buildHomelabStatus(metrics map[string]float64) *models.HomelabStatus {
// 	status := &models.HomelabStatus{}

// 	// CPU metrics
// 	if _, ok := metrics["node_cpu_seconds_total{cpu=0,mode=idle}"]; ok {
// 		// Count CPU cores
// 		cores := 0
// 		for key := range metrics {
// 			if strings.HasPrefix(key, "node_cpu_seconds_total") && strings.Contains(key, "mode=idle") {
// 				cores++
// 			}
// 		}
// 		status.CPU.Cores = cores

// 		// Calculate CPU usage
// 		var totalIdle, totalBusy float64
// 		for key, value := range metrics {
// 			if strings.HasPrefix(key, "node_cpu_seconds_total") {
// 				if strings.Contains(key, "mode=idle") {
// 					totalIdle += value
// 				} else {
// 					totalBusy += value
// 				}
// 			}
// 		}
// 		if totalIdle+totalBusy > 0 {
// 			status.CPU.UsagePercent = (totalBusy / (totalIdle + totalBusy)) * 100
// 		}
// 	}

// 	// Temperature (if available)
// 	for key, value := range metrics {
// 		if strings.Contains(key, "node_hwmon_temp_celsius") && strings.Contains(key, "chip=") {
// 			status.CPU.Temperature = value
// 			break
// 		}
// 	}

// 	// Memory metrics
// 	if memTotal, ok := metrics["node_memory_MemTotal_bytes"]; ok {
// 		status.Memory.Total = uint64(memTotal)
// 	}
// 	if memAvail, ok := metrics["node_memory_MemAvailable_bytes"]; ok {
// 		status.Memory.Available = uint64(memAvail)
// 		status.Memory.Used = status.Memory.Total - status.Memory.Available
// 		if status.Memory.Total > 0 {
// 			status.Memory.UsedPercent = (float64(status.Memory.Used) / float64(status.Memory.Total)) * 100
// 		}
// 	}

// 	// Disk metrics (root filesystem)
// 	for key, value := range metrics {
// 		if strings.Contains(key, "node_filesystem_size_bytes") && (strings.Contains(key, `mountpoint="/"`) || strings.Contains(key, "fstype=ext4")) {
// 			status.Disk.Total = uint64(value)
// 			// Find corresponding available bytes
// 			availKey := strings.Replace(key, "size_bytes", "avail_bytes", 1)
// 			if avail, ok := metrics[availKey]; ok {
// 				status.Disk.Available = uint64(avail)
// 				status.Disk.Used = status.Disk.Total - status.Disk.Available
// 				if status.Disk.Total > 0 {
// 					status.Disk.UsedPercent = (float64(status.Disk.Used) / float64(status.Disk.Total)) * 100
// 				}
// 			}
// 			break
// 		}
// 	}

// 	// Network metrics
// 	for key, value := range metrics {
// 		if strings.Contains(key, "node_network_receive_bytes_total") && !strings.Contains(key, "device=lo") {
// 			status.Network.BytesReceived += uint64(value)
// 		}
// 		if strings.Contains(key, "node_network_transmit_bytes_total") && !strings.Contains(key, "device=lo") {
// 			status.Network.BytesSent += uint64(value)
// 		}
// 	}

// 	// System uptime
// 	if uptime, ok := metrics["node_boot_time_seconds"]; ok {
// 		status.Uptime = uptime
// 	}

// 	// System info
// 	if nodename, ok := metrics["node_uname_info{nodename}"]; ok {
// 		status.SystemInfo.Hostname = strconv.FormatFloat(nodename, 'f', -1, 64)
// 	}

// 	// Try to extract hostname from metric labels
// 	for key := range metrics {
// 		if strings.HasPrefix(key, "node_uname_info") {
// 			if start := strings.Index(key, "nodename="); start != -1 {
// 				start += len("nodename=")
// 				end := strings.IndexAny(key[start:], ",}")
// 				if end != -1 {
// 					status.SystemInfo.Hostname = strings.Trim(key[start:start+end], `"`)
// 				}
// 			}
// 			if start := strings.Index(key, "sysname="); start != -1 {
// 				start += len("sysname=")
// 				end := strings.IndexAny(key[start:], ",}")
// 				if end != -1 {
// 					status.SystemInfo.OS = strings.Trim(key[start:start+end], `"`)
// 				}
// 			}
// 			if start := strings.Index(key, "release="); start != -1 {
// 				start += len("release=")
// 				end := strings.IndexAny(key[start:], ",}")
// 				if end != -1 {
// 					status.SystemInfo.Kernel = strings.Trim(key[start:start+end], `"`)
// 				}
// 			}
// 		}
// 	}

// 	return status
// }
