/*
Copyright 2019 The Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	metrics.Registry.Register(robotAccountExpiry)
}
