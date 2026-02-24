package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	httpSubsystem = "http"
)

var (
	// HTTPRequestsTotal counts total HTTP requests by method, path, and status code.
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: httpSubsystem,
			Name:      "requests_total",
			Help:      "Total number of HTTP requests processed.",
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTPRequestDuration tracks the latency of HTTP requests in seconds.
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: httpSubsystem,
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// HTTPRequestsInFlight tracks the number of in-flight HTTP requests.
	HTTPRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: httpSubsystem,
			Name:      "requests_in_flight",
			Help:      "Number of HTTP requests currently being processed.",
		},
	)

	// HTTPResponseSizeBytes tracks the size of HTTP responses.
	HTTPResponseSizeBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: httpSubsystem,
			Name:      "response_size_bytes",
			Help:      "Size of HTTP responses in bytes.",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 7), // 100B to 100MB
		},
		[]string{"method", "path"},
	)
)
