package metrics

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/log"
)

type debugMetricsEngine struct {
	logger log.Logger
}

var _ metricsEngine = (*debugMetricsEngine)(nil)

func NewDebugMetrics() Metrics {
	logger, err := createDebugLogger()
	if err != nil {
		panic(err)
	}
	engine := &debugMetricsEngine{logger: logger}
	return newMetrics(engine)
}

func (m *debugMetricsEngine) Start() {}

func (m *debugMetricsEngine) Stop() {}

func (m *debugMetricsEngine) recordRMWFailure(count uint64) {
	m.logger.Debug("recordRMWFailure", "count", count)
}

func (m *debugMetricsEngine) recordRMWInvalidated(count uint64) {
	m.logger.Debug("recordRMWInvalidated", "count", count)
}

func (m *debugMetricsEngine) recordForcedPreemption(count uint64) {
	m.logger.Debug("recordForcedPreemption", "count", count)
}

func (m *debugMetricsEngine) recordWakeupMiss(count uint64) {
	m.logger.Debug("recordWakeupMiss", "count", count)
}

func createDebugLogger() (log.Logger, error) {
	file, err := os.OpenFile("cannon-metrics.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(file.Name())
	if err != nil {
		return nil, err
	}

	fmt.Printf("Cannon debug metrics file will be saved to: %s\n", absPath)

	logger := log.NewLogger(log.LogfmtHandlerWithLevel(file, log.LevelDebug))
	return logger, nil
}
