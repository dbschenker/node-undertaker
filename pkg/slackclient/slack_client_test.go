package slackclient

import (
	"context"
	"github.com/dbschenker/node-undertaker/pkg/nodeundertaker/config"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSendNotification(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		webhookURL     string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Successful notification",
			message:        "Test message",
			webhookURL:     "/success",
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "Failed notification",
			message:        "Test message",
			webhookURL:     "/failure",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:           "Invalid URL",
			message:        "Test message",
			webhookURL:     "://invalid-url",
			expectedStatus: 0,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/success" {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			defer server.Close()

			webhookUrl, err := url.Parse(server.URL + tt.webhookURL)
			assert.NoError(t, err)

			cfg := &config.Config{
				NotificationsSlackWebhook: webhookUrl,
			}

			err = SendNotification(context.Background(), cfg, tt.message)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
