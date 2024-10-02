package metrics

type Metrics interface {
	TrackLLOp(step uint64, overwritesExistingReservation bool)
	TrackSCOpSuccess(step uint64)
	TrackSCOpFailure()
	TrackLLReservationInvalidated()
}

type baseMetricsImpl struct {
	lastLLOpStep uint64
}

var _ Metrics = (*baseMetricsImpl)(nil)

func NewNoopMetrics() Metrics {
	return &baseMetricsImpl{
		lastLLOpStep: 0,
	}
}

func (m *baseMetricsImpl) TrackLLOp(step uint64, overwritesExistingReservation bool) {
	if overwritesExistingReservation {
		m.recordRMWOverwritten()
	}
	m.lastLLOpStep = step
}

func (m *baseMetricsImpl) TrackSCOpSuccess(step uint64) {
	totalSteps := step - m.lastLLOpStep
	m.recordRMWSuccess(totalSteps)
}

func (m *baseMetricsImpl) TrackSCOpFailure() {
	m.recordRMWFailure()
}

func (m *baseMetricsImpl) TrackLLReservationInvalidated() {
	m.recordRMWInvalidated()
}

// TODO(#12061) Override the following methods in a derived struct for the production metrics implementation

func (n *baseMetricsImpl) recordRMWSuccess(totalSteps uint64) {}

func (n *baseMetricsImpl) recordRMWFailure() {}

func (n *baseMetricsImpl) recordRMWInvalidated() {}

func (n *baseMetricsImpl) recordRMWOverwritten() {}
