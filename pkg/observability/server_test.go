package observability

import (
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/observability/health"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLivenessServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(health.LivenessProbe))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	response, err := io.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, "{\"Healthy\":true}", string(response))
}

func TestReadinessServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(health.ReadinessProbe))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	response, err := io.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, "{\"Ready\":true}", string(response))
}

func TestMetricsServer(t *testing.T) {
	//dummy metric initialization
	var AppStartCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "mytopic",
			Subsystem: "mysystem",
			Name:      "myapp",
			Help:      "Number of starts for this app",
		},
	)
	prometheus.MustRegister(AppStartCounter)

	ts := httptest.NewServer(promhttp.Handler())
	defer ts.Close()

	res, err := http.Get(ts.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	response, err := io.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err)
	arrStrings := strings.Split(string(response), "\n")
	assert.Contains(t, arrStrings, "mytopic_mysystem_myapp 0")

}
