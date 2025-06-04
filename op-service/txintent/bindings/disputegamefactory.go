package bindings

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type DisputeGameFactoryFactory struct {
	BaseCallFactory
}

func newDisputeGameFactoryFactory(opts ...CallFactoryOption) *DisputeGameFactoryFactory {
	return &DisputeGameFactoryFactory{BaseCallFactory: *NewBaseCallFactory(opts...)}
}

type DisputeGameFactory struct {
	DisputeGameFactoryFactory

	Create    func(gameType uint32, claim eth.Bytes32, extraData []byte) TypedCall[common.Address] `sol:"create"`
	GameImpls func(gameType uint32) TypedCall[common.Address]                                      `sol:"gameImpls"`
	InitBonds func(gameType uint32) TypedCall[eth.ETH]                                             `sol:"initBonds"`
}

func NewDisputeGameFactory(opts ...CallFactoryOption) *DisputeGameFactory {
	factory := newDisputeGameFactoryFactory(opts...)
	impl := DisputeGameFactory{DisputeGameFactoryFactory: *factory}
	InitImpl(&impl)
	return &impl
}

type DisputeGameCreated struct {
	DisputeProxy common.Address
	GameType     uint32
	RootClaim    common.Hash
}

var DisputeGameCreatedSignature = crypto.Keccak256Hash([]byte("DisputeGameCreated(address,uint32,bytes32)"))

// TODO - fit log decoding into framework
func (d *DisputeGameFactory) DecodeDisputeGameCreatedLogs(receipt *types.Receipt) ([]DisputeGameCreated, error) {
	targetAddr, err := d.To()
	if err != nil {
		return nil, err
	}

	var logs []DisputeGameCreated
	for i := 0; i < len(receipt.Logs); i++ {
		log := receipt.Logs[i]
		// Check log is emitted from this contract
		if targetAddr != nil && log.Address != *targetAddr {
			continue
		}
		// Check that log has the expected signature
		if log.Topics[0] != DisputeGameCreatedSignature {
			continue
		}

		// We found a target log, parse the data out
		disputeProxy := common.HexToAddress(log.Topics[1].Hex())
		gameType := uint32(log.Topics[2].Big().Uint64())
		rootClaim := log.Topics[3]

		logs = append(logs, DisputeGameCreated{
			DisputeProxy: disputeProxy,
			GameType:     gameType,
			RootClaim:    rootClaim,
		})
	}

	return logs, nil
}
