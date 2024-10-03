package metrics

type Metrics interface {
	TrackLLOp(step uint64, overwritesExistingReservation bool)
	TrackSCOpSuccess(step uint64)
	TrackSCOpFailure()
	TrackLLReservationInvalidated()
	TrackPreemption(stepsSinceLastPreemption uint64)
}

type baseMetricsImpl struct {
	lastLLOpStep                    uint64
	rmwSuccessCount                 uint64
	rmwFailureCount                 uint64
	rmwReservationInvalidationCount uint64
	rmwReservationOverwriteCount    uint64
}

var _ Metrics = (*baseMetricsImpl)(nil)

func NewNoopMetrics() Metrics {
	impl := newBaseMetrics()
	return &impl
}

func newBaseMetrics() baseMetricsImpl {
	return baseMetricsImpl{
		lastLLOpStep:                    0,
		rmwSuccessCount:                 0,
		rmwFailureCount:                 0,
		rmwReservationInvalidationCount: 0,
		rmwReservationOverwriteCount:    0,
	}
}

func (m *baseMetricsImpl) TrackLLOp(step uint64, overwritesExistingReservation bool) {
	if overwritesExistingReservation {
		m.rmwReservationOverwriteCount += 1
		m.recordRMWOverwritten(m.rmwReservationOverwriteCount)
	}
	m.lastLLOpStep = step
}

func (m *baseMetricsImpl) TrackSCOpSuccess(step uint64) {
	m.rmwSuccessCount += 1
	totalSteps := step - m.lastLLOpStep
	m.recordRMWSuccess(m.rmwSuccessCount, totalSteps)
}

func (m *baseMetricsImpl) TrackSCOpFailure() {
	m.rmwFailureCount += 1
	m.recordRMWFailure(m.rmwFailureCount)
}

func (m *baseMetricsImpl) TrackLLReservationInvalidated() {
	m.rmwReservationInvalidationCount += 1
	m.recordRMWInvalidated(m.rmwReservationInvalidationCount)
}

func (m *baseMetricsImpl) TrackPreemption(stepsSinceLastPreemption uint64) {
	m.recordPreemption(stepsSinceLastPreemption)
}

// TODO(#12061) Override or implement the following for the production metrics implementation

func (n *baseMetricsImpl) recordRMWSuccess(count uint64, totalSteps uint64) {}

func (n *baseMetricsImpl) recordRMWFailure(count uint64) {}

func (n *baseMetricsImpl) recordRMWInvalidated(count uint64) {}

func (n *baseMetricsImpl) recordRMWOverwritten(count uint64) {}

func (n *baseMetricsImpl) recordPreemption(stepsSinceLastPreemption uint64) {}
