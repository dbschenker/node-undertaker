package node

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"strings"
	"time"
)

const (
	ReportingController = "dbschenker.com/node-undertaker"
)

func ReportEvent(ctx context.Context, cfg *config.Config, lvl log.Level, n NODE, action, reason, reasonDesc, msgOverride string) {
	microTime := metav1.NewMicroTime(time.Now())
	msg := msgOverride
	if msg == "" {
		if reasonDesc != "" {
			msg = fmt.Sprintf("%s/%s: %s %s due to %s", n.GetKind(), n.GetName(), strings.ToTitle(action), strings.ToLower(reason), reasonDesc)
		} else {
			msg = fmt.Sprintf("%s/%s: %s %s", n.GetKind(), n.GetName(), strings.ToTitle(action), strings.ToLower(reason))
		}
	}
	var eventType string
	switch lvl {
	case log.ErrorLevel:
		eventType = "Error"
	case log.WarnLevel:
		eventType = "Warning"
	case log.InfoLevel:
		eventType = "Normal"
	default:
		log.Errorf("Unsupported event level: %s", log.ErrorLevel.String())
		return
	}
	evt := eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("node-undertaker.%s", rand.String(16)),
			Namespace: cfg.Namespace,
		},
		Regarding: v1.ObjectReference{
			Namespace: cfg.Namespace,
			Name:      n.GetName(),
			Kind:      n.GetKind(),
		},
		Action:    action,
		Type:      eventType,
		Reason:    reason,
		EventTime: microTime,
		//Related: - second object related to event
		ReportingController: ReportingController,
		ReportingInstance:   cfg.Hostname,
		Note:                msg,
	}

	log.StandardLogger().Log(lvl, msg)
	_, err := cfg.K8sClient.EventsV1().Events(cfg.Namespace).Create(ctx, &evt, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("Couldn't create event: %s\n due to %v", msg, err)
	}
}
