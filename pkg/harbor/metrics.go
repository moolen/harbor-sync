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
	robotAccountExpiry = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "harbor_robot_account_expiry",
		Help: "The date after which the robot account expires. Expressed as Unix Epoch Time",
	}, []string{"project", "robot"})
)

func init() {
	metrics.Registry.Register(harborAPIRequestsHistogram)
}
