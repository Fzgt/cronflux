// Package metrics defines the Prometheus collectors cronflux exposes and wires
// them into a dedicated registry.
package metrics

import "github.com/prometheus/client_golang/prometheus"

// Metrics bundles every collector cronflux publishes. Construct it with New so
// the collectors are registered exactly once.
type Metrics struct {
	registry *prometheus.Registry
}

// New creates a Metrics backed by its own registry.
func New() *Metrics {
	return &Metrics{registry: prometheus.NewRegistry()}
}

// Registry exposes the underlying registry so the HTTP layer can serve it via
// promhttp.
func (m *Metrics) Registry() *prometheus.Registry { return m.registry }
