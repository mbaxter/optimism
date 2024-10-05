package metrics

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/log"
)

const versionTag = "cannon_version"

type influxMetricsEngine struct {
	influxClient *InfluxClient
	// Used to apply a cannon_version tag to all metrics
	cannonVersion string
	logger        log.Logger
}

var _ metricsEngine = (*influxMetricsEngine)(nil)

func NewInfluxMetrics(config InfluxConfig, cannonVersion string, logger log.Logger) Metrics {
	influxClient := NewInfluxClient(config, logger)
	engine := &influxMetricsEngine{influxClient: influxClient, cannonVersion: cannonVersion, logger: logger}
	return newMetrics(engine)
}

func (m *influxMetricsEngine) recordRMWSuccess(count uint64, totalSteps uint64) {
	err := m.influxClient.PushMetrics([]InfluxMetric{
		{
			Measurement: "cannon_rmw_success_count",
			Value:       count,
			Tags: map[string]string{
				versionTag: m.cannonVersion,
			},
		},
		{
			Measurement: "cannon_rmw_steps",
			Value:       totalSteps,
			Tags: map[string]string{
				versionTag: m.cannonVersion,
			},
		},
	})
	if err != nil {
		m.logger.Error("Failed to push metrics", "err", err)
	}
}

func (m *influxMetricsEngine) recordRMWFailure(count uint64) {
	err := m.influxClient.PushMetrics([]InfluxMetric{
		{
			Measurement: "cannon_rmw_failure_count",
			Value:       count,
			Tags: map[string]string{
				versionTag: m.cannonVersion,
			},
		},
	})
	if err != nil {
		m.logger.Error("Failed to push metrics", "err", err)
	}
}

func (m *influxMetricsEngine) recordRMWInvalidated(count uint64) {
	err := m.influxClient.PushMetrics([]InfluxMetric{
		{
			Measurement: "cannon_rmw_reset_count",
			Value:       count,
			Tags: map[string]string{
				versionTag: m.cannonVersion,
				"reason":   "invalidated",
			},
		},
	})
	if err != nil {
		m.logger.Error("Failed to push metrics", "err", err)
	}
}

func (m *influxMetricsEngine) recordRMWOverwritten(count uint64) {
	err := m.influxClient.PushMetrics([]InfluxMetric{
		{
			Measurement: "cannon_rmw_reset_count",
			Value:       count,
			Tags: map[string]string{
				versionTag: m.cannonVersion,
				"reason":   "overwritten",
			},
		},
	})
	if err != nil {
		m.logger.Error("Failed to push metrics", "err", err)
	}
}

func (m *influxMetricsEngine) recordPreemption(stepsSinceLastPreemption uint64) {
	m.logger.Info("Record preemption")
	err := m.influxClient.PushMetrics([]InfluxMetric{
		{
			Measurement: "cannon_step_count_at_preemption",
			Value:       stepsSinceLastPreemption,
			Tags: map[string]string{
				versionTag: m.cannonVersion,
			},
		},
	})
	if err != nil {
		m.logger.Error("Failed to push metrics", "err", err)
	}
}

type InfluxClient struct {
	config InfluxConfig
	logger log.Logger
}

type InfluxConfig struct {
	URL    string
	UserId string
	Token  string
}

type InfluxMetric struct {
	Measurement string
	Tags        map[string]string
	Value       uint64
}

func NewInfluxClient(config InfluxConfig, logger log.Logger) *InfluxClient {
	return &InfluxClient{config: config, logger: logger}
}

func (c *InfluxClient) PushMetrics(metrics []InfluxMetric) error {
	payload := []byte(c.createLineProtocolPayload(metrics))
	buffer := bytes.NewBuffer(payload)
	c.logger.Debug("Push metrics payload", "payload", payload)

	authToken := fmt.Sprintf("Bearer %s:%s", c.config.UserId, c.config.Token)
	request, err := http.NewRequest(http.MethodPost, c.config.URL, buffer)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "text/plain")
	request.Header.Set("Authorization", authToken)

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

	c.logger.Debug("Got metrics response", "status", response.StatusCode)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx response: %d %s", response.StatusCode, response.Status)
	}
	return nil
}

// createLineProtocolPayload creates a line protocol payload for a set of metrics
func (c *InfluxClient) createLineProtocolPayload(metrics []InfluxMetric) string {
	formatted := make([]string, len(metrics))
	for i, metric := range metrics {
		formatted[i] = fmt.Sprintf("%s%s metric=%d", metric.Measurement, c.fmtTags(metric.Tags), metric.Value)
	}
	return strings.Join(formatted, "\n")
}

// fmtTags formats a map of labels into a comma-separated string of key=value pairs
func (c *InfluxClient) fmtTags(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	formattedLabels := make([]string, 0)
	for k, v := range labels {
		formattedLabels = append(formattedLabels, fmt.Sprintf("%s=%s", k, v))
	}
	return "," + strings.Join(formattedLabels, ",")
}
