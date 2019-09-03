package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	matchingProjectsGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "harbor_matching_projects",
		Help: "The total number of matching projects per HarborSyncConfig",
	}, []string{"config", "selector_type", "selector_project_name"})
)

func init() {
	metrics.Registry.Register(matchingProjectsGauge)
}
