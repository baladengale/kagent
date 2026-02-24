package metrics

import (
	"testing"

	dto "github.com/prometheus/client_model/go"
)

func TestReconcileMetricsLabels(t *testing.T) {
	controllers := []string{"agent", "mcp_server", "remote_mcp_server", "mcp_service", "model_config", "model_provider_config"}

	for _, controller := range controllers {
		ReconcileTotal.WithLabelValues(controller, "success").Inc()
		ReconcileTotal.WithLabelValues(controller, "error").Inc()
		ReconcileErrors.WithLabelValues(controller).Inc()
		ReconcileDuration.WithLabelValues(controller).Observe(1.0)
	}

	// Verify a sample counter value
	var m dto.Metric
	counter := ReconcileTotal.WithLabelValues("agent", "success")
	if err := counter.Write(&m); err != nil {
		t.Fatalf("failed to write metric: %v", err)
	}

	if m.GetCounter().GetValue() < 1 {
		t.Errorf("expected counter value >= 1, got %f", m.GetCounter().GetValue())
	}
}

func TestManagedAgentsGauge(t *testing.T) {
	ManagedAgents.Set(5)

	var m dto.Metric
	if err := ManagedAgents.Write(&m); err != nil {
		t.Fatalf("failed to write metric: %v", err)
	}

	if m.GetGauge().GetValue() != 5 {
		t.Errorf("expected gauge value 5, got %f", m.GetGauge().GetValue())
	}
}

func TestToolDiscoveryMetrics(t *testing.T) {
	ToolDiscoveryDuration.WithLabelValues("test-server").Observe(0.1)
	ToolDiscoveryErrors.WithLabelValues("test-server").Inc()

	var m dto.Metric
	counter := ToolDiscoveryErrors.WithLabelValues("test-server")
	if err := counter.Write(&m); err != nil {
		t.Fatalf("failed to write metric: %v", err)
	}

	if m.GetCounter().GetValue() < 1 {
		t.Errorf("expected counter value >= 1, got %f", m.GetCounter().GetValue())
	}
}
