package cloudproviders

import "context"

//go:generate mockgen -destination=./mocks/api_mocks.go github.com/dbschenker/node-undertaker/pkg/cloudproviders CLOUDPROVIDER

type CLOUDPROVIDER interface {
	ValidateConfig() error

	// TerminateNode terminates node with provided providerId. Returns message (for creation of events) and error
	TerminateNode(context.Context, string) (string, error)
	// PrepareTermination prepares node to be termianted (i.e. removes it from load balancers)
	PrepareTermination(context.Context, string) (string, error)
}
