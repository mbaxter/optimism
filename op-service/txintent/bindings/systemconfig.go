package bindings

import (
	"github.com/ethereum/go-ethereum/common"
)

type SystemConfigFactory struct {
	BaseCallFactory
}

func newSystemConfigFactory(opts ...CallFactoryOption) *SystemConfigFactory {
	return &SystemConfigFactory{BaseCallFactory: *NewBaseCallFactory(opts...)}
}

type SystemConfig struct {
	SystemConfigFactory

	OptimismPortal     func() TypedCall[common.Address] `sol:"optimismPortal"`
	DisputeGameFactory func() TypedCall[common.Address] `sol:"disputeGameFactory"`
}

func NewSystemConfig(opts ...CallFactoryOption) *SystemConfig {
	factory := newSystemConfigFactory(opts...)
	impl := SystemConfig{SystemConfigFactory: *factory}
	InitImpl(&impl)
	return &impl
}
