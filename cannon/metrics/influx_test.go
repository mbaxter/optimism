package metrics

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

func TestInfluxClient_PushMetrics(t *testing.T) {
	influxConfig := InfluxConfig{
		URL:    "Defined Below",
		UserId: "123",
		Token:  "456",
	}
	expectedAuthHeader := fmt.Sprintf("Bearer %v:%v", influxConfig.UserId, influxConfig.Token)

	testCases := []struct {
		name            string
		metrics         []InfluxMetric
		statusCode      int
		expectedPayload string
		shouldError     bool
	}{
		{name: "single metric", statusCode: 204, metrics: []InfluxMetric{{Measurement: "a", Value: 11}}, expectedPayload: "a metric=11"},
		{name: "single metric with 1 tag", statusCode: 200, metrics: []InfluxMetric{{Measurement: "a", Tags: map[string]string{"t1": "some-tag"}, Value: 11}}, expectedPayload: "a,t1=some-tag metric=11"},
		{name: "single metric with multiple tags", statusCode: 299, metrics: []InfluxMetric{{Measurement: "a", Tags: map[string]string{"t1": "some-tag", "t2": "tag2"}, Value: 11}}, expectedPayload: "a,t1=some-tag,t2=tag2 metric=11"},
		{name: "two metrics", statusCode: 250, metrics: []InfluxMetric{{Measurement: "a", Value: 11}, {Measurement: "b", Tags: map[string]string{"t1": "tag"}, Value: 12}}, expectedPayload: "a metric=11\nb,t1=tag metric=12"},
		{name: "three metrics", statusCode: 201, metrics: []InfluxMetric{{Measurement: "a", Value: 11}, {Measurement: "b", Value: 12}, {Measurement: "c", Value: 13}}, expectedPayload: "a metric=11\nb metric=12\nc metric=13"},
		{name: "fails with status code 300", statusCode: 300, shouldError: true, metrics: []InfluxMetric{{Measurement: "a", Value: 11}}, expectedPayload: "a metric=11"},
		{name: "fails with status < 200", statusCode: 101, shouldError: true, metrics: []InfluxMetric{{Measurement: "a", Value: 11}}, expectedPayload: "a metric=11"},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			// Create a mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request method
				require.Equal(t, http.MethodPost, r.Method)

				// Verify the request headers
				authHeader := r.Header.Get("Authorization")
				require.Equal(t, expectedAuthHeader, authHeader)
				contentType := r.Header.Get("Content-Type")
				require.Equal(t, contentType, "text/plain")

				// Verify the request body
				b, err := io.ReadAll(r.Body)
				require.Nil(t, err)
				body := string(b)
				require.Equal(t, body, c.expectedPayload)

				// Send response
				w.WriteHeader(c.statusCode)
			}))
			defer server.Close()

			influxConfig.URL = server.URL
			client := NewInfluxClient(influxConfig, createLogger())

			// Call the PushMetric function
			err := client.PushMetrics(c.metrics)
			if c.shouldError {
				require.ErrorContains(t, err, "received non-2xx response")
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func createLogger() log.Logger {
	return log.NewLogger(log.LogfmtHandlerWithLevel(os.Stdout, log.LevelInfo))
}
