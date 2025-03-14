package node

import (
	"context"
	"fmt"
	"github.com/dbschenker/node-undertaker/pkg/nodeundertaker/config"
	"github.com/dbschenker/node-undertaker/pkg/slackclient"
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
			msg = fmt.Sprintf("%s due to %s", strings.ToLower(reason), reasonDesc)
		} else {
			msg = strings.ToLower(reason)
		}
	}

	fullMsg := msg

	if len(msg) >= 1024 {
		msg = msg[:1024]
	}

	var eventType string = ""
	switch lvl {
	case log.ErrorLevel:
		eventType = "Warning"
	case log.WarnLevel:
		eventType = "Warning"
	case log.InfoLevel:
		eventType = "Normal"
	default:
		log.Errorf("Unsupported event level: %s", log.ErrorLevel.String())
		return
	}
	evt := eventsv1.Event{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("node-undertaker.%s", rand.String(16)),
			Namespace: cfg.Namespace,
		},
		EventTime: microTime,
		//Related: - second object related to event
		ReportingController: ReportingController,
		ReportingInstance:   cfg.Hostname,
		Action:              action,
		Reason:              reason,
		Regarding: v1.ObjectReference{
			Namespace: cfg.Namespace,
			Name:      n.GetName(),
			Kind:      n.GetKind(),
		},

		Note: msg,
		Type: eventType,
	}

	log.StandardLogger().Log(lvl, fmt.Sprintf("%s/%s: %s", n.GetKind(), n.GetName(), fullMsg))
	_, err := cfg.K8sClient.EventsV1().Events(cfg.Namespace).Create(ctx, &evt, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("Couldn't create event: %s\n due to %v", msg, err)
	}
	if cfg.NotificationsSlackWebhook != nil {
		go func() {
			ctxWebhook, ctxWebhookCancel := context.WithTimeout(ctx, 10*time.Second)
			defer ctxWebhookCancel()
			err := slackclient.SendNotification(ctxWebhook, cfg, fmt.Sprintf("%s: %s/%s: %s", strings.ToUpper(lvl.String()), n.GetKind(), n.GetName(), msg))
			if err != nil {
				log.Errorf("Error sending notification to slack webhook")
			}
		}()
	}
}
