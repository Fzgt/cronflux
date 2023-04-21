// Package metrics defines the Prometheus collectors cronflux exposes and wires
// them into a dedicated registry.
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/Fzgt/cronflux/internal/job"
)

// Metrics bundles every collector cronflux publishes. Construct it with New so
// the collectors are registered exactly once.
type Metrics struct {
	registry *prometheus.Registry

	runsTotal    *prometheus.CounterVec
	retriesTotal *prometheus.CounterVec
	runDuration  *prometheus.HistogramVec
	ticksTotal   prometheus.Counter
	pendingRuns  prometheus.Gauge
	schedulerLag prometheus.Gauge
}

// New creates a Metrics backed by its own registry with all collectors
// registered.
func New() *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),
		runsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cronflux_runs_total",
			Help: "Number of runs that reached a terminal state, by job and state.",
		}, []string{"job", "state"}),
		retriesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cronflux_run_retries_total",
			Help: "Number of run retries scheduled, by job.",
		}, []string{"job"}),
		runDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "cronflux_run_duration_seconds",
			Help:    "Wall-clock duration of run executions, by job.",
			Buckets: prometheus.DefBuckets,
		}, []string{"job"}),
		ticksTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cronflux_scheduler_ticks_total",
			Help: "Number of scheduler ticks processed.",
		}),
		pendingRuns: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cronflux_pending_runs",
			Help: "Current number of runs waiting to be executed.",
		}),
		schedulerLag: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cronflux_scheduler_lag_seconds",
			Help: "Age of the oldest due-but-unclaimed run, in seconds.",
		}),
	}
	m.registry.MustRegister(
		m.runsTotal, m.retriesTotal, m.runDuration,
		m.ticksTotal, m.pendingRuns, m.schedulerLag,
	)
	return m
}

// Registry exposes the underlying registry so the HTTP layer can serve it via
// promhttp.
func (m *Metrics) Registry() *prometheus.Registry { return m.registry }

// ObserveRun records a terminal run outcome and its duration.
func (m *Metrics) ObserveRun(jobID string, state job.RunState, dur time.Duration) {
	m.runsTotal.WithLabelValues(jobID, string(state)).Inc()
	m.runDuration.WithLabelValues(jobID).Observe(dur.Seconds())
}

// IncRetry records that a retry was scheduled for a job.
func (m *Metrics) IncRetry(jobID string) {
	m.retriesTotal.WithLabelValues(jobID).Inc()
}

// Tick records a scheduler tick.
func (m *Metrics) Tick() { m.ticksTotal.Inc() }

// SetPending publishes the current pending-run backlog.
func (m *Metrics) SetPending(n int) { m.pendingRuns.Set(float64(n)) }

// SetLag publishes the age of the oldest unclaimed run.
func (m *Metrics) SetLag(d time.Duration) { m.schedulerLag.Set(d.Seconds()) }
