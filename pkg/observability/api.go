package observability

import (
	"context"
)

//go:generate mockgen -destination=./mocks/api_mocks.go gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/observability OBSERVABILITYSERVER

type OBSERVABILITYSERVER interface {
	StartServer(context.Context) error
	SetupRoutes()
}
