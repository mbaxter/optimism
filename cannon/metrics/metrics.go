package metrics

type Metrics interface {
	TrackLLOp(step uint64)
	TrackSCOpFailure()
	TrackLLReservationInvalidated()
	TrackForcedPreemption()
}

type metricsEngine interface {
	recordRMWFailure(count uint64)
	recordRMWInvalidated(count uint64)
	recordForcedPreemption(count uint64)
}

type metricsImpl struct {
	engine                          metricsEngine
	lastLLOpStep                    uint64
	rmwFailureCount                 uint64
	rmwReservationInvalidationCount uint64
	forcedPreemptionCount           uint64
}

var _ Metrics = (*metricsImpl)(nil)

func newMetrics(engine metricsEngine) Metrics {
	return &metricsImpl{
		engine:                          engine,
		lastLLOpStep:                    0,
		rmwFailureCount:                 0,
		rmwReservationInvalidationCount: 0,
		forcedPreemptionCount:           0,
	}
}

func (m *metricsImpl) TrackLLOp(step uint64) {
	m.lastLLOpStep = step
}

func (m *metricsImpl) TrackSCOpFailure() {
	m.rmwFailureCount += 1
	m.engine.recordRMWFailure(m.rmwFailureCount)
}

func (m *metricsImpl) TrackLLReservationInvalidated() {
	m.rmwReservationInvalidationCount += 1
	m.engine.recordRMWInvalidated(m.rmwReservationInvalidationCount)
}

func (m *metricsImpl) TrackForcedPreemption() {
	m.forcedPreemptionCount += 1
	m.engine.recordForcedPreemption(m.forcedPreemptionCount)
}
