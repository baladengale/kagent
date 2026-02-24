package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	databaseSubsystem = "database"
)

var (
	// DatabaseOperationsTotal counts total database operations by operation type and table.
	DatabaseOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: databaseSubsystem,
			Name:      "operations_total",
			Help:      "Total number of database operations.",
		},
		[]string{"operation", "table"},
	)

	// DatabaseOperationDuration tracks the duration of database operations in seconds.
	DatabaseOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: databaseSubsystem,
			Name:      "operation_duration_seconds",
			Help:      "Duration of database operations in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// DatabaseErrors counts database operation errors.
	DatabaseErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: databaseSubsystem,
			Name:      "errors_total",
			Help:      "Total number of database operation errors.",
		},
		[]string{"operation", "table"},
	)
)
