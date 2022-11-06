package txpool

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
	"math/big"
	"os"
	"testing"
)

func NewCalcBaseFee(config *params.ChainConfig, parent *types.Header) *big.Int {
	// If the current block is the first EIP-1559 block, return the InitialBaseFee.
	if !config.IsLondon(parent.Number) {
		return new(big.Int).SetUint64(params.InitialBaseFee)
	}

	parentGasTarget := parent.GasLimit / params.ElasticityMultiplier
	// If the parent gasUsed is the same as the target, the baseFee remains unchanged.
	if parent.GasUsed == parentGasTarget {
		return new(big.Int).Set(parent.BaseFee)
	}

	var (
		num   = new(big.Int)
		denom = new(big.Int)
	)

	testFactor := uint64(2)
	if parent.GasUsed > parentGasTarget {
		// If the parent block used more gas than its target, the baseFee should increase.
		// max(1, parentBaseFee * gasUsedDelta / parentGasTarget / baseFeeChangeDenominator)
		num.SetUint64(parent.GasUsed - parentGasTarget)
		num.Mul(num, parent.BaseFee)
		num.Div(num, denom.SetUint64(parentGasTarget))
		num.Div(num, denom.SetUint64(params.BaseFeeChangeDenominator*testFactor))
		baseFeeDelta := math.BigMax(num, common.Big1)

		return num.Add(parent.BaseFee, baseFeeDelta)
	} else {
		// Otherwise if the parent block used less gas than its target, the baseFee should decrease.
		// max(0, parentBaseFee * gasUsedDelta / parentGasTarget / baseFeeChangeDenominator)
		num.SetUint64(parentGasTarget - parent.GasUsed)
		num.Mul(num, parent.BaseFee)
		num.Div(num, denom.SetUint64(parentGasTarget))
		num.Div(num, denom.SetUint64(params.BaseFeeChangeDenominator*testFactor))
		baseFee := num.Sub(parent.BaseFee, num)

		return math.BigMax(baseFee, common.Big0)
	}
}

func TestBaseFeeCalc(t *testing.T) {

	BorMainnetChainConfig := &params.ChainConfig{
		ChainID:             big.NewInt(137),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      true,
		EIP150Hash:          common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(3395000),
		MuirGlacierBlock:    big.NewInt(3395000),
		BerlinBlock:         big.NewInt(14750000),
		LondonBlock:         big.NewInt(23850000),
	}

	ec, err := ethclient.Dial("https://polygon-rpc.com")
	require.NoError(t, err)
	defer ec.Close()

	f, err := os.Create("/Users/pauloleary/work/base-fee-calc.tsv")
	require.NoError(t, err)
	defer f.Close()

	var newFeeCalc *big.Int
	for i := int64(34621784); i < 34921784; i++ {
		blk, err := ec.BlockByNumber(context.Background(), big.NewInt(i))
		if err != nil {
			continue
		}

		h := blk.Header()
		checkFeeCalc := misc.CalcBaseFee(BorMainnetChainConfig, h)

		if newFeeCalc != nil {
			h.BaseFee = newFeeCalc
		}
		newFeeCalc = NewCalcBaseFee(BorMainnetChainConfig, h)

		s := fmt.Sprintf("%v\t%.1f\t%.1f\n", i,
			float64(checkFeeCalc.Int64())/1_000_000_000,
			float64(newFeeCalc.Int64())/1_000_000_000)

		f.Write([]byte(s))
	}
}
