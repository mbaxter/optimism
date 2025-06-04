package bindings

import (
	"math/big"
)

type AnchorStateRegistryFactory struct {
	BaseCallFactory
}

func newAnchorStateRegistryFactory(opts ...CallFactoryOption) *AnchorStateRegistryFactory {
	return &AnchorStateRegistryFactory{BaseCallFactory: *NewBaseCallFactory(opts...)}
}

type AnchorRoot struct {
	Hash     [32]byte
	L2SeqNum *big.Int
}

type AnchorStateRegistry struct {
	AnchorStateRegistryFactory

	RespectedGameType func() TypedCall[uint32]     `sol:"respectedGameType"`
	GetAnchorRoot     func() TypedCall[AnchorRoot] `sol:"getAnchorRoot"`
}

func NewAnchorStateRegistry(opts ...CallFactoryOption) *AnchorStateRegistry {
	factory := newAnchorStateRegistryFactory(opts...)
	impl := AnchorStateRegistry{AnchorStateRegistryFactory: *factory}
	InitImpl(&impl)
	return &impl
}
