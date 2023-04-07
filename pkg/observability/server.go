package observability

import (
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/observability/health"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/observability/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"net/http"
)

func StartServer(config *config.Config) error {

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/livez", health.LivenessProbe)
	http.HandleFunc("readyz", health.ReadinessProbe)
	hostAddress := fmt.Sprintf(":%v", config.Port)
	return http.ListenAndServe(hostAddress, nil)
}

func init() {
	metrics.Initialize()
}
