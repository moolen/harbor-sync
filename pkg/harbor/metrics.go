package harbor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	harborAPIRequestsHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of the last http request",
	}, []string{"code", "method", "path"})
)

func init() {
	metrics.Registry.Register(harborAPIRequestsHistogram)
}
