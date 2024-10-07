package metrics

func NewNoopMetrics() Metrics {
	return newMetrics(&noopMetricsEngine{})
}

type noopMetricsEngine struct{}

var _ metricsEngine = (*noopMetricsEngine)(nil)

func (n noopMetricsEngine) recordRMWFailure(count uint64) {}

func (n noopMetricsEngine) recordRMWInvalidated(count uint64) {}

func (n noopMetricsEngine) recordForcedPreemption(uint64) {}

func (n noopMetricsEngine) recordWakeupMiss(count uint64) {}
