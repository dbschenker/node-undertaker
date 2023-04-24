package nodeupdatehandler

import (
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestNodeIsGrownUp(t *testing.T) {
	cfg := config.Config{NodeInitialThreshold: 5}
	creationTime := metav1.Now().Add(-20 * time.Second)

	node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "dummy",
			CreationTimestamp: metav1.NewTime(creationTime),
		},
	}

	res := nodeIsGrownUp(&cfg, &node)
	assert.True(t, res)
}

func TestNodeIsGrownUpNot(t *testing.T) {
	cfg := config.Config{NodeInitialThreshold: 90}
	creationTime := metav1.Now()

	node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "dummy",
			CreationTimestamp: creationTime,
		},
	}

	res := nodeIsGrownUp(&cfg, &node)
	assert.False(t, res)
}
