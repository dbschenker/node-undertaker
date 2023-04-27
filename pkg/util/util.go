package util

import (
	"context"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/kubeclient"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	log "github.com/sirupsen/logrus"
)

func ProcessingMessage(ctx context.Context, cfg *config.Config, lvl log.Level, msg string) {
	switch lvl {
	case log.ErrorLevel:
		log.Error(msg)
	case log.WarnLevel:
		log.Warn(msg)
	case log.InfoLevel:
		log.Info(msg)
	case log.DebugLevel:
		log.Debug(msg)
	}
	err := kubeclient.CreateEvent(ctx, cfg, lvl, msg)
	if err != nil {
		log.Errorf("Coundn't create event due to: %v", err)
	}
}
