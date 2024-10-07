package metrics

type ServiceState int

const (
	Idle ServiceState = iota
	Running
	Stopped
)

type Service interface {
	Start()
	Stop()
}
