package kwok

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k KwokCloudProvider) CreateNode(ctx context.Context, name string) error {

	node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"node.alpha.kubernetes.io/ttl": "0",
				"kwok.x-k8s.io/node":           "fake",
			},
			Labels: map[string]string{
				"beta.kubernetes.io/arch":       "amd64",
				"beta.kubernetes.io/os":         "linux",
				"kubernetes.io/arch":            "amd64",
				"kubernetes.io/hostname":        name,
				"kubernetes.io/os":              "linux",
				"kubernetes.io/role":            "agent",
				"node-role.kubernetes.io/agent": "",
				"type":                          "kwok",
			},
		},
		Spec: v1.NodeSpec{
			ProviderID: fmt.Sprintf("kwok://%s", name),
		},
		Status: v1.NodeStatus{
			Allocatable: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse("32"),
				v1.ResourceMemory: resource.MustParse("64Gi"),
				v1.ResourcePods:   resource.MustParse("100"),
			},
			Capacity: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse("32"),
				v1.ResourceMemory: resource.MustParse("64Gi"),
				v1.ResourcePods:   resource.MustParse("100"),
			},
			NodeInfo: v1.NodeSystemInfo{
				Architecture:            "amd64",
				BootID:                  "",
				ContainerRuntimeVersion: "",
				KernelVersion:           "",
				KubeProxyVersion:        "fake",
				KubeletVersion:          "fake",
				MachineID:               "",
				OperatingSystem:         "linux",
				SystemUUID:              "",
				OSImage:                 "",
			},
		},
	}

	_, err := k.K8sClient.CoreV1().Nodes().Create(ctx, &node, metav1.CreateOptions{})

	return err
}
