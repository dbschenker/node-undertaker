package health

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io"
	"net/http/httptest"
	"testing"
)

func TestLivenessProbe(t *testing.T) {
	expectedResponse := liveness{Healthy: true}

	req := httptest.NewRequest("GET", "http://lcoalhost:8081/livez", nil)
	w := httptest.NewRecorder()
	LivenessProbe(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var response liveness
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Equal(t, expectedResponse, response)
}

func TestReadinessProbe(t *testing.T) {
	expectedResponse := readiness{Ready: true}

	req := httptest.NewRequest("GET", "http://lcoalhost:8081/readyz", nil)
	w := httptest.NewRecorder()
	ReadinessProbe(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var response readiness
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Equal(t, expectedResponse, response)
}
