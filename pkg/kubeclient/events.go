package kubeclient

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func CreateEvent(ctx context.Context, cfg *config.Config, lvl log.Level, msg string) error {
	microTime := metav1.NewMicroTime(time.Now())
	var eventType string
	switch lvl {
	case log.ErrorLevel:
		eventType = "Error"
	case log.WarnLevel:
		eventType = "Warning"
	case log.InfoLevel:
		eventType = "Normal"
	case log.DebugLevel:
		eventType = "Normal"
	}
	evt := eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("node-undertaker.%d", microTime.UnixMicro()),
			Namespace: cfg.Namespace,
		},
		Regarding: v1.ObjectReference{
			Namespace: cfg.Namespace,
			Name:      "test",
			Kind:      "nodes",
		},
		Action:    "test",
		Type:      eventType,
		Reason:    "TODO",
		EventTime: microTime,
		//Regarding:  - object  related to the event
		//Related: - second object related to event
		//EventTime: metav1.NewMicroTime(time.Now()),
		ReportingController: "dbschenker.com/node-undertaker",
		ReportingInstance:   "TODO", //fg.K8sClient.
		Note:                msg,
	}

	_, err := cfg.K8sClient.EventsV1().Events(cfg.Namespace).Create(ctx, &evt, metav1.CreateOptions{})
	return err
}
