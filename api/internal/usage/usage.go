package usage

import (
	"api/internal/metrics"
	"log"
	"runtime"
	"time"
)

func MonitorMemory(registry *metrics.Registry) {
	log.Println("[USAGE] Memory monitor goroutine started")
	log.Println("[USAGE] Monitoring interval: 20 seconds")
	
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	iteration := 0
	for range ticker.C {
		iteration++
		log.Printf("[USAGE] Memory check iteration #%d", iteration)
		
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		log.Printf("[USAGE] Memory stats - Alloc: %d bytes (%.2f MB)", m.Alloc, float64(m.Alloc)/1024/1024)
		log.Printf("[USAGE] Memory stats - TotalAlloc: %d bytes (%.2f MB)", m.TotalAlloc, float64(m.TotalAlloc)/1024/1024)
		log.Printf("[USAGE] Memory stats - Sys: %d bytes (%.2f MB)", m.Sys, float64(m.Sys)/1024/1024)
		log.Printf("[USAGE] Memory stats - HeapAlloc: %d bytes (%.2f MB)", m.HeapAlloc, float64(m.HeapAlloc)/1024/1024)
		log.Printf("[USAGE] Memory stats - HeapInuse: %d bytes (%.2f MB)", m.HeapInuse, float64(m.HeapInuse)/1024/1024)
		log.Printf("[USAGE] Memory stats - NumGC: %d", m.NumGC)
		log.Printf("[USAGE] Memory stats - NumGoroutine: %d", runtime.NumGoroutine())

		registry.SetGauge("app_memory_usage_bytes", float64(m.Alloc), map[string]string{"type": "alloc"})
		registry.SetGauge("app_memory_usage_bytes", float64(m.TotalAlloc), map[string]string{"type": "total_alloc"})
		registry.SetGauge("app_memory_usage_bytes", float64(m.Sys), map[string]string{"type": "sys"})
		registry.SetGauge("app_memory_usage_bytes", float64(m.HeapAlloc), map[string]string{"type": "heap_alloc"})
		registry.SetGauge("app_memory_usage_bytes", float64(m.HeapInuse), map[string]string{"type": "heap_inuse"})
		registry.SetGauge("app_gc_runs_total", float64(m.NumGC), map[string]string{})
		
		log.Printf("[USAGE] Metrics updated successfully")
	}
}

func GetMemoryUsage() map[string]uint64 {
	log.Println("[USAGE] Getting current memory usage snapshot")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	usage := map[string]uint64{
		"alloc":       m.Alloc,
		"total_alloc": m.TotalAlloc,
		"sys":         m.Sys,
		"num_gc":      uint64(m.NumGC),
	}
	
	log.Printf("[USAGE] Memory snapshot: %+v", usage)
	return usage
}
