package models

type HomelabStatus struct {
	CPU        CPUMetrics
	Memory     MemoryMetrics
	Disk       DiskMetrics
	Network    NetworkMetrics
	SystemInfo SystemInfo
	Uptime     float64 // in seconds
}

type CPUMetrics struct {
	UsagePercent float64
	Cores        int
	Temperature  float64 // in Celsius
}

type MemoryMetrics struct {
	Total       uint64 // in bytes
	Used        uint64 // in bytes
	UsedPercent float64
	Available   uint64 // in bytes
}

type DiskMetrics struct {
	Total       uint64 // in bytes
	Used        uint64 // in bytes
	UsedPercent float64
	Available   uint64 // in bytes
}

type NetworkMetrics struct {
	BytesSent     uint64
	BytesReceived uint64
}

type SystemInfo struct {
	Hostname string
	OS       string
	Kernel   string
}
