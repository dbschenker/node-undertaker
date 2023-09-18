package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/client-go/listers/core/v1"
)

func Initialize(lister v1.NodeLister) func() {
	nsc := CreateNodeStatusCollector(lister)
	prometheus.MustRegister(nsc)

	return func() {
		prometheus.Unregister(nsc)
	}
}
