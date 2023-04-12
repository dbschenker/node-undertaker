package nodehealthnotificationhandler

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	log "github.com/sirupsen/logrus"
)

type DefaultNodeHealthNotificationHandler struct {
}

func (d DefaultNodeHealthNotificationHandler) HandleHealthMessages(ctx context.Context, cfg *config.Config) error {
	errCounter := 0
	for {
		messages, err := cfg.CloudProvider.ReceiveMessages(ctx)
		if err != nil {
			log.Errorf("Got error when trying to receive messages: %v", err)
			errCounter += 1
			if errCounter > 10 {
				return fmt.Errorf("couldn't handle message %d times", errCounter)
			}
		} else {
			errCounter = 0
		}
		for i := range messages {
			go handleHealthMessage(ctx, cfg, messages[i])
		}
	}
}

func handleHealthMessage(ctx context.Context, cfg *config.Config, message cloudproviders.Message) {
	// annotate that node is unhealthy
	err := cfg.CloudProvider.CompleteMessage(ctx, message)
	if err != nil {
		log.Errorf("Couldn't complete message: %v because of: %v", message, err)
	}
	// drain
	// annotate
	// delete from cloud provider
	// annotate
	// leave it for cloud-controller to delete from k8s cluster
}
