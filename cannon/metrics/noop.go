package metrics

type NoopMetrics struct {
}

var _ Metrics = (*Metrics)(nil)

func NewNoopMetrics() *NoopMetrics {
	return &NoopMetrics{}
}
