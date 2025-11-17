package usage

import (
	"api/internal/metrics"
	"runtime"
	"time"
)

func MonitorMemory(registry *metrics.Registry) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		registry.SetGauge("app_memory_usage_bytes", float64(m.Alloc), map[string]string{"type": "alloc"})
		registry.SetGauge("app_memory_usage_bytes", float64(m.TotalAlloc), map[string]string{"type": "total_alloc"})
		registry.SetGauge("app_memory_usage_bytes", float64(m.Sys), map[string]string{"type": "sys"})
	}
}

func GetMemoryUsage() map[string]uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]uint64{
		"alloc":       m.Alloc,
		"total_alloc": m.TotalAlloc,
		"sys":         m.Sys,
		"num_gc":      uint64(m.NumGC),
	}
}
