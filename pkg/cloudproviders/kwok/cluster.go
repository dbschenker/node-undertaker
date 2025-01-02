package kwok

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"testing"
)

const (
	KwokImage = "registry.k8s.io/kwok/cluster:v0.6.1-k8s.v1.29.8"
	KwokPort  = "8080"
)

func StartCluster(t *testing.T, ctx context.Context) (*kubernetes.Clientset, error) {
	t.Helper()

	port := fmt.Sprintf("%s/tcp", KwokPort)

	req := testcontainers.ContainerRequest{
		Image:        KwokImage,
		ExposedPorts: []string{port},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port(port)),
			wait.ForHTTP("/readyz").WithPort(nat.Port(port)),
		),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, KwokPort)
	if err != nil {
		return nil, err
	}

	k8sCfg := rest.Config{
		Host: fmt.Sprintf("%s:%s", ip, mappedPort.Port()),
	}

	clientset, err := kubernetes.NewForConfig(&k8sCfg)

	return clientset, err
}
