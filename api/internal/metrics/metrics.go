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

	counterCount := 0
	for name, counters := range r.counters {
		output.WriteString(fmt.Sprintf("# HELP %s Total number of requests\n", name))
		output.WriteString(fmt.Sprintf("# TYPE %s counter\n", name))
		for _, counter := range counters {
			output.WriteString(fmt.Sprintf("%s%s %g\n", name, formatLabels(counter.labels), counter.value))
			counterCount++
		}
	}

	gaugeCount := 0
	for name, gauges := range r.gauges {
		output.WriteString(fmt.Sprintf("# HELP %s Current memory usage\n", name))
		output.WriteString(fmt.Sprintf("# TYPE %s gauge\n", name))
		for _, gauge := range gauges {
			output.WriteString(fmt.Sprintf("%s%s %g\n", name, formatLabels(gauge.labels), gauge.value))
			gaugeCount++
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
