package metrics

import (
	"github.com/dbschenker/node-undertaker/pkg/nodeundertaker/node"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/client-go/listers/core/v1"
)

const (
	NodeHealthyLabelOverride = "healthy"
	MetricsNamespace         = "node_undertaker"
	NodeMetricsSubsystem     = "node"
	HealthMetricName         = "health"
	MetricLabelNode          = "node"
	MetricLabelStatus        = "status"
)

type NodeStatusCollector struct {
	lister v1.NodeLister
	desc   *prometheus.Desc
}

func CreateNodeStatusCollector(lister v1.NodeLister) NodeStatusCollector {

	ret := NodeStatusCollector{
		lister: lister,
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricsNamespace, NodeMetricsSubsystem, HealthMetricName),
			"Node health status",
			[]string{MetricLabelNode, MetricLabelStatus}, nil,
		),
	}
	return ret
}

func (nsc NodeStatusCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- nsc.desc
	//prometheus.DescribeByCollect(nsc, descs)
}

func (nsc NodeStatusCollector) Collect(metrics chan<- prometheus.Metric) {
	nodes, err := nsc.lister.List(labels.Everything())
	if err != nil {
		log.Errorf("Error while collecting metrics: %v", err)
		return
	}
	for i := range nodes {
		n := node.CreateNode(nodes[i])
		statusLabel := n.GetLabel()
		if statusLabel == node.NodeHealthy {
			statusLabel = NodeHealthyLabelOverride
		}
		metrics <- prometheus.MustNewConstMetric(nsc.desc, prometheus.GaugeValue, 1.0, n.GetName(), statusLabel)

	}
}
