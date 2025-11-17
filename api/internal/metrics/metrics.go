package metrics

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

type Counter struct {
	value  float64
	labels map[string]string
}

type Gauge struct {
	value  float64
	labels map[string]string
}

type Registry struct {
	counters map[string][]*Counter
	gauges   map[string][]*Gauge
	mu       sync.RWMutex
}

func NewRegistry() *Registry {
	log.Println("[METRICS] Creating new metrics registry")
	return &Registry{
		counters: make(map[string][]*Counter),
		gauges:   make(map[string][]*Gauge),
	}
}

func (r *Registry) IncrementCounter(name string, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("[METRICS] Incrementing counter '%s' with labels %v", name, labels)

	for _, counter := range r.counters[name] {
		if labelsMatch(counter.labels, labels) {
			counter.value++
			log.Printf("[METRICS] Counter '%s' incremented to %.0f", name, counter.value)
			return
		}
	}

	newCounter := &Counter{
		value:  1,
		labels: labels,
	}
	r.counters[name] = append(r.counters[name], newCounter)
	log.Printf("[METRICS] New counter '%s' created with value 1", name)
}

func (r *Registry) SetGauge(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("[METRICS] Setting gauge '%s' to %.2f with labels %v", name, value, labels)

	for _, gauge := range r.gauges[name] {
		if labelsMatch(gauge.labels, labels) {
			gauge.value = value
			return
		}
	}

	newGauge := &Gauge{
		value:  value,
		labels: labels,
	}
	r.gauges[name] = append(r.gauges[name], newGauge)
	log.Printf("[METRICS] New gauge '%s' created with value %.2f", name, value)
}

func (r *Registry) Export() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	log.Println("[METRICS] Exporting metrics in Prometheus format")
	var output strings.Builder

	allCounters := map[string]map[string]string{
		"api_requests_total": {
			"help": "Total number of API requests by method, endpoint and status",
		},
		"redis_operations_total": {
			"help": "Total number of Redis operations by operation type and status",
		},
		"redis_load_test_runs_total": {
			"help": "Total number of Redis load test runs by status",
		},
		"db_load_test_runs_total": {
			"help": "Total number of database load test runs by status",
		},
	}

	counterCount := 0
	for name, meta := range allCounters {
		output.WriteString(fmt.Sprintf("# HELP %s %s\n", name, meta["help"]))
		output.WriteString(fmt.Sprintf("# TYPE %s counter\n", name))
		
		if counters, exists := r.counters[name]; exists {
			for _, counter := range counters {
				output.WriteString(fmt.Sprintf("%s%s %g\n", name, formatLabels(counter.labels), counter.value))
				counterCount++
			}
		} else {
			output.WriteString(fmt.Sprintf("%s 0\n", name))
		}
	}

	allGauges := map[string]map[string]string{
		"app_memory_usage_bytes": {
			"help": "Current application memory usage in bytes by type",
		},
		"app_goroutines": {
			"help": "Current number of goroutines running in the application",
		},
		"app_gc_runs_total": {
			"help": "Total number of garbage collection runs",
		},
		"redis_connection_status": {
			"help": "Redis connection status (1=connected, 0=disconnected)",
		},
		"postgres_connection_status": {
			"help": "PostgreSQL connection status (1=connected, 0=disconnected)",
		},
		"redis_load_test_duration_seconds": {
			"help": "Duration of the last Redis load test in seconds",
		},
		"redis_load_test_throughput_keys_per_sec": {
			"help": "Throughput of the last Redis load test in keys per second",
		},
		"redis_load_test_successful_keys": {
			"help": "Number of successfully loaded keys in the last Redis load test",
		},
		"redis_load_test_failed_keys": {
			"help": "Number of failed keys in the last Redis load test",
		},
		"redis_load_test_total_bytes": {
			"help": "Total bytes written in the last Redis load test",
		},
		"app_loaded_keys_count": {
			"help": "Number of keys currently stored in application memory from load test",
		},
		"app_loaded_values_count": {
			"help": "Number of values currently stored in application memory from load test",
		},
		"http_request_duration_seconds": {
			"help": "HTTP request duration in seconds by endpoint",
		},
		"redis_operation_latency_seconds": {
			"help": "Redis operation latency in seconds by operation type",
		},
		"postgres_operation_latency_seconds": {
			"help": "PostgreSQL operation latency in seconds by operation type",
		},
		"db_load_test_successful_connections": {
			"help": "Number of successful database connections in the last load test",
		},
		"db_load_test_failed_connections": {
			"help": "Number of failed database connections in the last load test",
		},
		"db_load_test_duration_seconds": {
			"help": "Duration of the last database load test in seconds",
		},
		"db_load_test_average_latency_seconds": {
			"help": "Average connection latency in the last database load test",
		},
		"db_active_connections_count": {
			"help": "Number of active database connections being kept alive",
		},
	}

	gaugeCount := 0
	for name, meta := range allGauges {
		output.WriteString(fmt.Sprintf("# HELP %s %s\n", name, meta["help"]))
		output.WriteString(fmt.Sprintf("# TYPE %s gauge\n", name))
		
		if gauges, exists := r.gauges[name]; exists {
			for _, gauge := range gauges {
				output.WriteString(fmt.Sprintf("%s%s %g\n", name, formatLabels(gauge.labels), gauge.value))
				gaugeCount++
			}
		} else {
			output.WriteString(fmt.Sprintf("%s 0\n", name))
		}
	}

	log.Printf("[METRICS] Exported %d counters and %d gauges", counterCount, gaugeCount)
	return output.String()
}

func labelsMatch(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
	}
	return "{" + strings.Join(parts, ",") + "}"
}
