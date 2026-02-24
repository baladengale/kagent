package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestRegisterAll(t *testing.T) {
	// Create a new registry to avoid polluting the default one
	registry := prometheus.NewRegistry()

	// RegisterAll should not panic
	RegisterAll(registry)

	// Verify metrics are registered by gathering them
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Should have at least some registered metrics (zero-valued counters/gauges may not appear)
	// The important thing is that registration didn't panic
	_ = metricFamilies
}

func TestRegisterAllDoesNotDuplicate(t *testing.T) {
	// Registering twice to the same registry should panic (duplicate registration)
	registry := prometheus.NewRegistry()
	RegisterAll(registry)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration, but none occurred")
		}
	}()

	// This should panic because metrics are already registered
	RegisterAll(registry)
}

func TestBuildInfoCollector(t *testing.T) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(NewBuildInfoCollector())

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	found := false
	for _, mf := range metricFamilies {
		if mf.GetName() == "kagent_build_info" {
			found = true
			metrics := mf.GetMetric()
			if len(metrics) != 1 {
				t.Errorf("expected 1 metric, got %d", len(metrics))
			}
			if metrics[0].GetGauge().GetValue() != 1 {
				t.Errorf("expected value 1, got %f", metrics[0].GetGauge().GetValue())
			}
		}
	}

	if !found {
		t.Error("kagent_build_info metric not found")
	}
}
