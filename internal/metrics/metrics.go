package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	BackupsByController = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "backups_operator_backups",
			Help: "Number of created backups",
		},
		[]string{"name", "namespace", "controller", "status"},
	)

	ScheduledTaskFailuresByControllerTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "backups_operator_scheduled_task_failures_total",
			Help: "Count of backup schedule failures",
		},
		[]string{"name", "namespace", "controller", "type"},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(
		BackupsByController,
		ScheduledTaskFailuresByControllerTotal,
	)
}
