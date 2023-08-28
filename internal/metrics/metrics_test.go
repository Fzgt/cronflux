package metrics_test

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/metrics"
)

// gaugeOrCounter finds a metric by name and label set and returns its value.
func gaugeOrCounter(t *testing.T, g prometheus.Gatherer, name string, labels map[string]string) float64 {
	t.Helper()
	mfs, err := g.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		for _, m := range mf.GetMetric() {
			if !labelsMatch(m.GetLabel(), labels) {
				continue
			}
			if c := m.GetCounter(); c != nil {
				return c.GetValue()
			}
			if gg := m.GetGauge(); gg != nil {
				return gg.GetValue()
			}
		}
	}
	t.Fatalf("metric %s%v not found", name, labels)
	return 0
}

func labelsMatch(pairs []*dto.LabelPair, want map[string]string) bool {
	got := map[string]string{}
	for _, p := range pairs {
		got[p.GetName()] = p.GetValue()
	}
	for k, v := range want {
		if got[k] != v {
			return false
		}
	}
	return true
}

func TestObserveRunAndRetry(t *testing.T) {
	m := metrics.New()
	m.ObserveRun("build", job.StateSucceeded, 1500*time.Millisecond)
	m.ObserveRun("build", job.StateSucceeded, 500*time.Millisecond)
	m.IncRetry("build")

	if v := gaugeOrCounter(t, m.Registry(), "cronflux_runs_total", map[string]string{"job": "build", "state": "succeeded"}); v != 2 {
		t.Errorf("runs_total = %v, want 2", v)
	}
	if v := gaugeOrCounter(t, m.Registry(), "cronflux_run_retries_total", map[string]string{"job": "build"}); v != 1 {
		t.Errorf("retries_total = %v, want 1", v)
	}
}

func TestTickAndGauges(t *testing.T) {
	m := metrics.New()
	m.Tick()
	m.Tick()
	m.SetPending(3)
	m.SetLag(5 * time.Second)

	if v := gaugeOrCounter(t, m.Registry(), "cronflux_scheduler_ticks_total", nil); v != 2 {
		t.Errorf("ticks_total = %v, want 2", v)
	}
	if v := gaugeOrCounter(t, m.Registry(), "cronflux_pending_runs", nil); v != 3 {
		t.Errorf("pending_runs = %v, want 3", v)
	}
	if v := gaugeOrCounter(t, m.Registry(), "cronflux_scheduler_lag_seconds", nil); v != 5 {
		t.Errorf("lag_seconds = %v, want 5", v)
	}
}
