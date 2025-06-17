package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ServiceOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_operations_total",
			Help: "Total number of service operations by type and status.",
		},
		[]string{"operation", "status"},
	)

	ServiceOperationDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "service_operation_duration_seconds",
			Help:    "Duration of service operations in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "status"},
	)

	TransferMonitorAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transfer_monitor_attempts_total",
			Help: "Total attempts made by the transfer monitor.",
		},
		[]string{"transfer_id", "attempt_number", "result"},
	)
)
