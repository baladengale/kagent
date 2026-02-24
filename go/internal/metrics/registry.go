package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const metricsNamespace = "kagent"

// RegisterAll registers all kagent custom metrics with the given registerer.
func RegisterAll(registerer prometheus.Registerer) {
	registerer.MustRegister(
		// HTTP metrics
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HTTPRequestsInFlight,
		HTTPResponseSizeBytes,

		// Reconciler metrics
		ReconcileTotal,
		ReconcileDuration,
		ReconcileErrors,
		ManagedAgents,
		ManagedToolServers,
		ToolDiscoveryDuration,
		ToolDiscoveryErrors,

		// Database metrics
		DatabaseOperationsTotal,
		DatabaseOperationDuration,
		DatabaseErrors,
	)
}
