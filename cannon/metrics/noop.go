package metrics

func NewNoopMetrics() Metrics {
	return newMetrics(&noopMetricsEngine{})
}

type noopMetricsEngine struct{}

var _ metricsEngine = (*noopMetricsEngine)(nil)

func (n noopMetricsEngine) recordRMWSuccess(count uint64, totalSteps uint64) {}

func (n noopMetricsEngine) recordRMWFailure(count uint64) {}

func (n noopMetricsEngine) recordRMWInvalidated(count uint64) {}

func (n noopMetricsEngine) recordRMWOverwritten(count uint64) {}

func (n noopMetricsEngine) recordPreemption(stepsSinceLastPreemption uint64) {}
