package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	reconcilerSubsystem = "reconciler"
)

var (
	// ReconcileTotal counts the total number of reconciliations by controller and result.
	ReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: reconcilerSubsystem,
			Name:      "reconcile_total",
			Help:      "Total number of reconciliations by controller and result.",
		},
		[]string{"controller", "result"},
	)

	// ReconcileDuration tracks the duration of reconciliations in seconds.
	ReconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: reconcilerSubsystem,
			Name:      "reconcile_duration_seconds",
			Help:      "Duration of reconciliations in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"controller"},
	)

	// ReconcileErrors counts reconciliation errors by controller and error type.
	ReconcileErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: reconcilerSubsystem,
			Name:      "reconcile_errors_total",
			Help:      "Total number of reconciliation errors by controller.",
		},
		[]string{"controller"},
	)

	// ManagedAgents tracks the current number of managed agents.
	ManagedAgents = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: reconcilerSubsystem,
			Name:      "managed_agents",
			Help:      "Current number of agents managed by the controller.",
		},
	)

	// ManagedToolServers tracks the current number of managed tool servers.
	ManagedToolServers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: reconcilerSubsystem,
			Name:      "managed_tool_servers",
			Help:      "Current number of tool servers managed by the controller.",
		},
	)

	// ToolDiscoveryDuration tracks the duration of MCP tool discovery operations.
	ToolDiscoveryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: reconcilerSubsystem,
			Name:      "tool_discovery_duration_seconds",
			Help:      "Duration of MCP tool discovery operations in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"server"},
	)

	// ToolDiscoveryErrors counts tool discovery errors.
	ToolDiscoveryErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: reconcilerSubsystem,
			Name:      "tool_discovery_errors_total",
			Help:      "Total number of MCP tool discovery errors.",
		},
		[]string{"server"},
	)
)
