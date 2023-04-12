package nodehealthnotificationhandler

import (
	"context"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
)

//go:generate mockgen -destination=./mocks/api_mocks.go gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodehealthnotificationhandler NODEHEALTHNOTIFICATIONHANDLER

type NODEHEALTHNOTIFICATIONHANDLER interface {
	HandleHealthMessages(context.Context, *config.Config) error
}
