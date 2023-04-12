package nodeprovider

import (
	"fmt"
	types "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/struct"
)

type K8sNodeProvider struct {
	DrainTimeout int
}

func (np K8sNodeProvider) GetNodeState(node types.Node) (NodeState, error) {
	return None, fmt.Errorf("TODO")
}

func (np K8sNodeProvider) SetNodeState(node types.Node, state NodeState) error {
	return fmt.Errorf("TODO")
}

func (np K8sNodeProvider) DrainNode(node types.Node) error {
	return fmt.Errorf("TODO")
}
