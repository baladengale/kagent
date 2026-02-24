package metrics

import (
	"testing"

	dto "github.com/prometheus/client_model/go"
)

func TestDatabaseMetricsLabels(t *testing.T) {
	operations := []string{"save", "get", "list", "delete"}
	tables := []string{"agents", "sessions", "tools", "tool_servers"}

	for _, op := range operations {
		for _, table := range tables {
			DatabaseOperationsTotal.WithLabelValues(op, table).Inc()
			DatabaseOperationDuration.WithLabelValues(op, table).Observe(0.01)
		}
	}

	// Verify a sample counter value
	var m dto.Metric
	counter := DatabaseOperationsTotal.WithLabelValues("save", "agents")
	if err := counter.Write(&m); err != nil {
		t.Fatalf("failed to write metric: %v", err)
	}

	if m.GetCounter().GetValue() < 1 {
		t.Errorf("expected counter value >= 1, got %f", m.GetCounter().GetValue())
	}
}

func TestDatabaseErrorMetrics(t *testing.T) {
	DatabaseErrors.WithLabelValues("save", "agents").Inc()

	var m dto.Metric
	counter := DatabaseErrors.WithLabelValues("save", "agents")
	if err := counter.Write(&m); err != nil {
		t.Fatalf("failed to write metric: %v", err)
	}

	if m.GetCounter().GetValue() < 1 {
		t.Errorf("expected counter value >= 1, got %f", m.GetCounter().GetValue())
	}
}
