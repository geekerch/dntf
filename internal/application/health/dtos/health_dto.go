package dtos

import "time"

type LivenessResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type DetailedHealthResponse struct {
	Status       string             `json:"status"`
	Timestamp    time.Time          `json:"timestamp"`
	Version      string             `json:"version"`
	Dependencies []DependencyHealth `json:"dependencies"`
	SystemInfo   SystemInfo         `json:"system_info"`
}

type DependencyHealth struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

type SystemInfo struct {
	Uptime    string     `json:"uptime"`
	GoVersion string     `json:"go_version"`
	Memory    MemoryInfo `json:"memory"`
}

type MemoryInfo struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
}

type LegacyHealthResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
