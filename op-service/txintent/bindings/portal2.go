package bindings

import (
	"github.com/ethereum/go-ethereum/common"
)

type Portal2Factory struct {
	BaseCallFactory
}

func newPortal2Factory(opts ...CallFactoryOption) *Portal2Factory {
	return &Portal2Factory{BaseCallFactory: *NewBaseCallFactory(opts...)}
}

type Portal2 struct {
	Portal2Factory

	AnchorStateRegistry func() TypedCall[common.Address] `sol:"anchorStateRegistry"`
}

func NewPortal2(opts ...CallFactoryOption) *Portal2 {
	factory := newPortal2Factory(opts...)
	impl := Portal2{Portal2Factory: *factory}
	InitImpl(&impl)
	return &impl
}
