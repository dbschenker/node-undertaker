package slackclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dbschenker/node-undertaker/pkg/nodeundertaker/config"
	"io"
	"net/http"
)

func SendNotification(ctx context.Context, cfg *config.Config, message string) error {

	payload := struct {
		Text string `json:"text"`
	}{
		Text: message,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error while marshalling slack notification: %v", err)
	}

	resp, err := http.Post(cfg.NotificationsSlackWebhook.String(), "application/json", bytes.NewBuffer(data))
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("error sending notification to slack endpoint - got %d http code", resp.StatusCode)
	}
	return err
}
