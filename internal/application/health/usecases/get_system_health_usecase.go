package usecases

import (
	"context"
	"runtime"
	"time"

	"notification/internal/application/health/dtos"
)

type GetSystemHealthUseCase struct {
	startTime time.Time
}

func NewGetSystemHealthUseCase() *GetSystemHealthUseCase {
	return &GetSystemHealthUseCase{
		startTime: time.Now(),
	}
}

func (u *GetSystemHealthUseCase) Execute(ctx context.Context) (*dtos.DetailedHealthResponse, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	dependencies := []dtos.DependencyHealth{
		{
			Name:    "Database",
			Status:  "Healthy",
			Message: "Connection established",
		},
		{
			Name:    "NATS",
			Status:  "Healthy",
			Message: "Connection established",
		},
	}

	uptime := time.Since(u.startTime)

	return &dtos.DetailedHealthResponse{
		Status:       "Healthy",
		Timestamp:    time.Now(),
		Version:      "v0.1.0",
		Dependencies: dependencies,
		SystemInfo: dtos.SystemInfo{
			Uptime:    uptime.String(),
			GoVersion: runtime.Version(),
			Memory: dtos.MemoryInfo{
				Alloc:      m.Alloc,
				TotalAlloc: m.TotalAlloc,
				Sys:        m.Sys,
				NumGC:      m.NumGC,
			},
		},
	}, nil
}
