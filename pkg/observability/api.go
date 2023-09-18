package observability

import (
	"context"
)

//go:generate mockgen -destination=./mocks/api_mocks.go github.com/dbschenker/node-undertaker/pkg/observability OBSERVABILITYSERVER

type OBSERVABILITYSERVER interface {
	StartServer(context.Context) error
	SetupRoutes()
}
