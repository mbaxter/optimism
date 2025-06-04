package proofs

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	gameTypes "github.com/ethereum-optimism/optimism/op-challenger/game/fault/types"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl/contract"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txintent/bindings"
	"github.com/ethereum-optimism/optimism/op-service/txplan"
)

var badClaim = eth.Bytes32(common.HexToHash("0xdeadbeef00000000000000000000000000000000000000000000000000000000"))

func TestChallengerPlaysGame(gt *testing.T) {
	// Setup
	t := devtest.SerialT(gt)
	sys := presets.NewSimpleInterop(t)
	sys.L1Network.WaitForOnline()

	attacker := fundAttackerWallet(t, sys, eth.OneEther.Mul(2))
	game := createNewGame(t, sys, attacker)
	t.Logf("Game created at %s", game.Hex())

	// TODO - wait for challenger to respond to game, play game down to split depth
}

func fundAttackerWallet(t devtest.T, sys *presets.SimpleInterop, fundingAmount eth.ETH) *dsl.EOA {
	wallet := sys.Wallet.NewEOA(sys.L1EL)
	initialBalance := sys.FunderL1.FundAtLeast(wallet, fundingAmount)
	require.GreaterOrEqual(t, initialBalance.ToBig().Int64(), fundingAmount.ToBig().Int64())

	return wallet
}

func createNewGame(t devtest.T, sys *presets.SimpleInterop, attacker *dsl.EOA) common.Address {
	l1Client := sys.L1EL.Escape().EthClient()
	dgfAddr := sys.L2ChainA.Escape().Deployment().DisputeGameFactoryProxyAddr()
	dgf := bindings.NewDisputeGameFactory(bindings.WithClient(l1Client), bindings.WithTo(dgfAddr), bindings.WithTest(t))

	// Pull some metadata we need to construct a new game
	gameType := uint32(gameTypes.SuperCannonGameType)
	anchorRoot := getAnchorRoot(t, sys)
	requiredBonds := contract.Read(dgf.InitBonds(gameType))

	l2SeqNum := big.NewInt(0).Add(anchorRoot.L2SeqNum, big.NewInt(10))
	extraData := common.BigToHash(l2SeqNum).Bytes()
	receipt := contract.Write(attacker, dgf.Create(gameType, badClaim, extraData), txplan.WithValue(requiredBonds.ToBig()))

	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)

	// Extract new game contract from the logs
	createLogs, err := dgf.DecodeDisputeGameCreatedLogs(receipt)
	require.NoError(t, err)
	require.Equal(t, 1, len(createLogs))

	return createLogs[0].DisputeProxy
}

func getAnchorRoot(t devtest.T, sys *presets.SimpleInterop) bindings.AnchorRoot {
	l1Client := sys.L1EL.Escape().EthClient()
	sysConfigAddr := sys.L2ChainA.Escape().Deployment().SystemConfigProxyAddr()
	sysConfig := bindings.NewSystemConfig(bindings.WithClient(l1Client), bindings.WithTo(sysConfigAddr), bindings.WithTest(t))

	portalAddr := contract.Read(sysConfig.OptimismPortal())
	portal := bindings.NewPortal2(bindings.WithClient(l1Client), bindings.WithTo(portalAddr), bindings.WithTest(t))

	asrAddr := contract.Read(portal.AnchorStateRegistry())
	asr := bindings.NewAnchorStateRegistry(bindings.WithClient(l1Client), bindings.WithTo(asrAddr), bindings.WithTest(t))

	return contract.Read(asr.GetAnchorRoot())
}
