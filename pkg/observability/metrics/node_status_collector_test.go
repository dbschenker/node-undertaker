package metrics

//go:generate mockgen -destination=./mocks/informer_mocks.go k8s.io/client-go/listers/core/v1 NodeLister

import (
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/node"
	mock_v1 "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/observability/metrics/mocks"
	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"testing"
)

func TestCreateNodeStatusCollectorDescribe(t *testing.T) {
	nsc := CreateNodeStatusCollector(nil)

	c := make(chan *prometheus.Desc)
	go nsc.Describe(c)
	desc := <-c
	assert.NotNil(t, desc)
}

func TestCreateNodeStatusCollectorCollect(t *testing.T) {
	nodes := []*v1.Node{
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
			},
		},
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-2",
			},
		},
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-3",
				Labels: map[string]string{
					node.Label: node.NodeDraining,
				},
			},
		},
	}
	mockCtrl := gomock.NewController(t)
	fakeLister := mock_v1.NewMockNodeLister(mockCtrl)
	fakeLister.EXPECT().List(gomock.Any()).Return(nodes, nil).Times(2)

	nsc := CreateNodeStatusCollector(fakeLister)

	const expectedMetadata = `
		# HELP node_undertaker_node_health Node health status
		# TYPE node_undertaker_node_health gauge
	`
	expectedMetricText := `
		node_undertaker_node_health{node="node-1",status="healthy"} 1
		node_undertaker_node_health{node="node-2",status="healthy"} 1
		node_undertaker_node_health{node="node-3",status="draining"} 1
	`

	count := testutil.CollectAndCount(nsc)
	assert.Equal(t, 3, count)
	err := testutil.CollectAndCompare(nsc, strings.NewReader(expectedMetadata+expectedMetricText))
	assert.NoError(t, err)
}
