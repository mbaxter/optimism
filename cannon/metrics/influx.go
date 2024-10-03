package metrics

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/log"
)

type influxMetrics struct {
	baseMetricsImpl
	category     string
	influxClient *InfluxClient
}

var _ Metrics = (*influxMetrics)(nil)

func NewInfluxMetrics(category string, config InfluxConfig, logger log.Logger) Metrics {
	influxClient := NewInfluxClient(config, logger)
	return &influxMetrics{baseMetricsImpl: newBaseMetrics(), category: category, influxClient: influxClient}
}

func (m *influxMetrics) recordRMWSuccess(count uint64, totalSteps uint64) {
	m.influxClient.PushMetric(Metric{
		Measurement: m.category,
		Fields: map[string]uint64{
			"rmw_success_count": count,
			"rmw_step_count":    totalSteps,
		},
	})
}

func (m *influxMetrics) recordRMWFailure(count uint64) {
	m.influxClient.PushMetric(Metric{
		Measurement: m.category,
		Fields: map[string]uint64{
			"rmw_failure_count": count,
		},
	})
}

func (m *influxMetrics) recordRMWInvalidated(count uint64) {
	m.influxClient.PushMetric(Metric{
		Measurement: m.category,
		Tags: map[string]string{
			"reason": "memory_invalidated",
		},
		Fields: map[string]uint64{
			"rmw_reset_count": count,
		},
	})
}

func (m *influxMetrics) recordRMWOverwritten(count uint64) {
	m.influxClient.PushMetric(Metric{
		Measurement: m.category,
		Tags: map[string]string{
			"reason": "overwritten",
		},
		Fields: map[string]uint64{
			"rmw_reset_count": count,
		},
	})
}

func (m *influxMetrics) recordPreemption(stepsSinceLastPreemption uint64) {
	m.influxClient.PushMetric(Metric{
		Measurement: m.category,
		Fields: map[string]uint64{
			"steps_at_preemption": stepsSinceLastPreemption,
		},
	})
}

type InfluxClient struct {
	config InfluxConfig
	logger log.Logger
}

type InfluxConfig struct {
	URL      string `json:"url"`
	User     string `json:"user"`
	Password string `json:"pass"`
}

type Metric struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]uint64
}

func NewInfluxClient(config InfluxConfig, logger log.Logger) *InfluxClient {
	return &InfluxClient{config: config, logger: logger}
}

func (c *InfluxClient) PushMetric(metric Metric) error {
	payload := []byte(c.createLineProtocolPayload(metric))
	buffer := bytes.NewBuffer(payload)

	request, err := http.NewRequest(http.MethodPost, c.config.URL, buffer)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "text/plain")
	request.Header.Set("Authorization", c.authToken())

	// Create an HTTP client and send the request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("Failed to close response body", "err", err)
		}
	}(response.Body)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx response: %d %s", response.StatusCode, response.Status)
	}
	return nil
}

// authToken creates a basic auth token for the given user and password
func (c *InfluxClient) authToken() string {
	// Add basic authentication header
	auth := c.config.User + ":" + c.config.Password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// createLineProtocolPayload creates a line protocol payload for a single metric
func (c *InfluxClient) createLineProtocolPayload(metric Metric) string {
	return fmt.Sprintf("%s,%s %s", metric.Measurement, c.fmtTags(metric.Tags), c.fmtFields(metric.Fields))
}

// fmtTags formats a map of labels into a comma-separated string of key=value pairs
func (c *InfluxClient) fmtTags(labels map[string]string) string {
	formattedLabels := make([]string, 0)
	for k, v := range labels {
		formattedLabels = append(formattedLabels, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(formattedLabels, ",")
}

// fmtTags formats a map of uint fields into a comma-separated string of key=value pairs
func (c *InfluxClient) fmtFields(fields map[string]uint64) string {
	formatted := make([]string, 0)
	for k, v := range fields {
		formatted = append(formatted, fmt.Sprintf("%s=%du", k, v))
	}
	return strings.Join(formatted, ",")
}
