package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestHTTPMetricsLabels(t *testing.T) {
	// Test that counters work with expected label values
	HTTPRequestsTotal.WithLabelValues("GET", "/api/agents", "200").Inc()

	var m dto.Metric
	counter := HTTPRequestsTotal.WithLabelValues("GET", "/api/agents", "200")
	if err := counter.Write(&m); err != nil {
		t.Fatalf("failed to write metric: %v", err)
	}

	if m.GetCounter().GetValue() < 1 {
		t.Errorf("expected counter value >= 1, got %f", m.GetCounter().GetValue())
	}
}

func TestHTTPRequestDuration(t *testing.T) {
	HTTPRequestDuration.WithLabelValues("GET", "/api/agents").Observe(0.5)

	var m dto.Metric
	observer := HTTPRequestDuration.WithLabelValues("GET", "/api/agents").(prometheus.Metric)
	if err := observer.Write(&m); err != nil {
		t.Fatalf("failed to write metric: %v", err)
	}

	if m.GetHistogram().GetSampleCount() < 1 {
		t.Errorf("expected at least 1 sample, got %d", m.GetHistogram().GetSampleCount())
	}
}

func TestHTTPRequestsInFlight(t *testing.T) {
	HTTPRequestsInFlight.Inc()
	HTTPRequestsInFlight.Dec()

	var m dto.Metric
	if err := HTTPRequestsInFlight.Write(&m); err != nil {
		t.Fatalf("failed to write metric: %v", err)
	}

	// After inc then dec, should be back to 0
	if m.GetGauge().GetValue() != 0 {
		t.Errorf("expected gauge value 0 after inc/dec, got %f", m.GetGauge().GetValue())
	}
}
