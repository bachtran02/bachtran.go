package models

type NodeStatus struct {
	CPU        CPUMetrics
	Memory     MemoryMetrics
	Disk       DiskMetrics
	Network    NetworkMetrics
	SystemInfo SystemInfo
	Online     bool
	Uptime     float64 // in seconds
}

type CPUMetrics struct {
	UsagePercent float64
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
	UpSpeed   float64
	DownSpeed float64
}

type SystemInfo struct {
	Hostname     string
	Cores        int
	OS           string
	MaxFreqGHz   float64
	Architecture string
	TotalRAM     uint64
	TotalDisk    uint64
}
