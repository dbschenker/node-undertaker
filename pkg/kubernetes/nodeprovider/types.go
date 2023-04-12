package nodeprovider

import (
	types "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/struct"
)

type NodeProvider interface {
	GetNodeState(node types.Node) (NodeState, error)
	SetNodeState(node types.Node, state NodeState) error
	DrainNode(node types.Node) error
}

type NodeState int

const (
	None NodeState = iota
	Unhealthy
	Draining
	Drained
	DrainFailed
	DeletedFromCloudProvider
)
