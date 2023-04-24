package cloudproviders

import "context"

type CloudProvider interface {
	ValidateConfig() error
	TerminateNode(context.Context, string) error
}
