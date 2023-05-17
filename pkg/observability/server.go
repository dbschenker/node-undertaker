package observability

import (
	"context"
	"errors"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/observability/health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type DefaultObservabilityServer struct {
	server *http.Server
}

func GetDefaultObservabilityServer(config *config.Config) DefaultObservabilityServer {
	o := DefaultObservabilityServer{}
	hostAddress := fmt.Sprintf(":%v", config.Port)
	o.server = &http.Server{
		Addr: hostAddress,
	}
	return o
}

func (o *DefaultObservabilityServer) SetupRoutes() {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/livez", health.LivenessProbe)
	http.HandleFunc("/readyz", health.ReadinessProbe)
}

func (o *DefaultObservabilityServer) StartServer(ctx context.Context) error {

	go func() {
		select {
		case <-ctx.Done():
			log.Debugf("shutting down prometheus server")
			o.server.Shutdown(ctx)
		}
	}()
	err := o.server.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
