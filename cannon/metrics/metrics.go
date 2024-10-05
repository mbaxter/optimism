package metrics

type Metrics interface {
	TrackLLOp(step uint64, overwritesExistingReservation bool)
	TrackSCOpSuccess(step uint64)
	TrackSCOpFailure()
	TrackLLReservationInvalidated()
	TrackPreemption(stepsSinceLastPreemption uint64)
}

type metricsEngine interface {
	recordRMWSuccess(count uint64, totalSteps uint64)
	recordRMWFailure(count uint64)
	recordRMWInvalidated(count uint64)
	recordRMWOverwritten(count uint64)
	recordPreemption(stepsSinceLastPreemption uint64)
}

type metricsImpl struct {
	engine                          metricsEngine
	lastLLOpStep                    uint64
	rmwSuccessCount                 uint64
	rmwFailureCount                 uint64
	rmwReservationInvalidationCount uint64
	rmwReservationOverwriteCount    uint64
}

var _ Metrics = (*metricsImpl)(nil)

func newMetrics(engine metricsEngine) Metrics {
	return &metricsImpl{
		engine:                          engine,
		lastLLOpStep:                    0,
		rmwSuccessCount:                 0,
		rmwFailureCount:                 0,
		rmwReservationInvalidationCount: 0,
		rmwReservationOverwriteCount:    0,
	}
}

func (m *metricsImpl) TrackLLOp(step uint64, overwritesExistingReservation bool) {
	if overwritesExistingReservation {
		m.rmwReservationOverwriteCount += 1
		m.engine.recordRMWOverwritten(m.rmwReservationOverwriteCount)
	}
	m.lastLLOpStep = step
}

func (m *metricsImpl) TrackSCOpSuccess(step uint64) {
	m.rmwSuccessCount += 1
	totalSteps := step - m.lastLLOpStep
	m.engine.recordRMWSuccess(m.rmwSuccessCount, totalSteps)
}

func (m *metricsImpl) TrackSCOpFailure() {
	m.rmwFailureCount += 1
	m.engine.recordRMWFailure(m.rmwFailureCount)
}

func (m *metricsImpl) TrackLLReservationInvalidated() {
	m.rmwReservationInvalidationCount += 1
	m.engine.recordRMWInvalidated(m.rmwReservationInvalidationCount)
}

func (m *metricsImpl) TrackPreemption(stepsSinceLastPreemption uint64) {
	m.engine.recordPreemption(stepsSinceLastPreemption)
}
