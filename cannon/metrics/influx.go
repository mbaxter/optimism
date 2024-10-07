package metrics

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

const versionTag = "cannon_version"

type influxMetricsEngine struct {
	influxClient   *InfluxClient
	cannonVersion  string
	batchSize      int
	pendingMetrics []InfluxMetric
	pushTimer      *time.Ticker
	stopChan       chan struct{}
	mutex          sync.Mutex
	logger         log.Logger
	state          ServiceState
}

var _ metricsEngine = (*influxMetricsEngine)(nil)

func NewInfluxMetrics(config InfluxConfig, cannonVersion string, logger log.Logger) Metrics {
	pushInterval := 1 * time.Second
	batchSize := 50
	pushTimer := time.NewTicker(pushInterval)
	influxClient := NewInfluxClient(config, logger)
	engine := &influxMetricsEngine{influxClient: influxClient, cannonVersion: cannonVersion, batchSize: batchSize, pushTimer: pushTimer, stopChan: make(chan struct{}), logger: logger}

	return newMetrics(engine)
}

func (m *influxMetricsEngine) Start() {
	if m.state == Idle {
		m.state = Running
		go m.periodicPush()
	}
}

func (m *influxMetricsEngine) Stop() {
	if m.state == Running {
		if len(m.pendingMetrics) != 0 {
			// Flush remaining data
			m.mutex.Lock()
			defer m.mutex.Unlock()
			m.push()
		}
		m.state = Stopped
		close(m.stopChan)
	}
}

func (m *influxMetricsEngine) recordMetric(metric InfluxMetric) {
	if m.state != Running {
		m.logger.Error("Must be running to record metrics", "state", m.state)
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.pendingMetrics = append(m.pendingMetrics, metric)

	// Flush if we've hit the batch size
	if len(m.pendingMetrics) >= m.batchSize {
		m.push()
	}
}

func (m *influxMetricsEngine) periodicPush() {
	for {
		select {
		case <-m.pushTimer.C:
			m.mutex.Lock()
			m.push()
			m.mutex.Unlock()
		case <-m.stopChan:
			m.pushTimer.Stop()
			return
		}
	}
}

func (m *influxMetricsEngine) push() {
	if len(m.pendingMetrics) == 0 {
		return
	}

	err := m.influxClient.PushMetrics(m.pendingMetrics)
	if err != nil {
		m.logger.Error("Failed to push metrics", "err", err)
	}

	m.pendingMetrics = []InfluxMetric{}
}

func (m *influxMetricsEngine) recordRMWFailure(count uint64) {
	m.recordMetric(InfluxMetric{
		Measurement: "cannon_rmw_failure_count",
		Value:       count,
		Tags: map[string]string{
			versionTag: m.cannonVersion,
		},
	})
}

func (m *influxMetricsEngine) recordRMWInvalidated(count uint64) {
	m.recordMetric(InfluxMetric{
		Measurement: "cannon_rmw_reset_count",
		Value:       count,
		Tags: map[string]string{
			versionTag: m.cannonVersion,
			"reason":   "invalidated",
		},
	})
}

func (m *influxMetricsEngine) recordForcedPreemption(count uint64) {
	m.recordMetric(InfluxMetric{
		Measurement: "cannon_forced_preemption_count",
		Value:       count,
		Tags: map[string]string{
			versionTag: m.cannonVersion,
		},
	})
}

func (m *influxMetricsEngine) recordWakeupMiss(count uint64) {
	m.recordMetric(InfluxMetric{
		Measurement: "cannon_wakeup_miss_count",
		Value:       count,
		Tags: map[string]string{
			versionTag: m.cannonVersion,
		},
	})
}

type InfluxClient struct {
	config     InfluxConfig
	logger     log.Logger
	httpClient *http.Client
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
	return &InfluxClient{config: config, logger: logger, httpClient: &http.Client{}}
}

func (c *InfluxClient) PushMetrics(metrics []InfluxMetric) error {
	payload := c.createLineProtocolPayload(metrics)
	authToken := fmt.Sprintf("Bearer %s:%s", c.config.UserId, c.config.Token)
	request, err := http.NewRequest(http.MethodPost, c.config.URL, payload)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "text/plain")
	request.Header.Set("Authorization", authToken)

	// Create an HTTP client and send the request
	response, err := c.httpClient.Do(request) //nolint:bodyclose
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
func (c *InfluxClient) createLineProtocolPayload(metrics []InfluxMetric) *bytes.Buffer {
	buf := new(bytes.Buffer)
	for _, metric := range metrics {
		buf.WriteString(fmt.Sprintf("%s%s metric=%d\n", metric.Measurement, c.fmtTags(metric.Tags), metric.Value))
	}
	return buf
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
