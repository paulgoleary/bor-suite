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
	"path/filepath"
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

	if parent.GasUsed > parentGasTarget {
		// If the parent block used more gas than its target, the baseFee should increase.
		// max(1, parentBaseFee * gasUsedDelta / parentGasTarget / baseFeeChangeDenominator)
		num.SetUint64(parent.GasUsed - parentGasTarget)
		num.Mul(num, parent.BaseFee)
		num.Div(num, denom.SetUint64(parentGasTarget))
		num.Div(num, denom.SetUint64(params.BaseFeeChangeDenominatorPostDelhi))
		baseFeeDelta := math.BigMax(num, common.Big1)

		return num.Add(parent.BaseFee, baseFeeDelta)
	} else {
		// Otherwise if the parent block used less gas than its target, the baseFee should decrease.
		// max(0, parentBaseFee * gasUsedDelta / parentGasTarget / baseFeeChangeDenominator)
		num.SetUint64(parentGasTarget - parent.GasUsed)
		num.Mul(num, parent.BaseFee)
		num.Div(num, denom.SetUint64(parentGasTarget))
		num.Div(num, denom.SetUint64(params.BaseFeeChangeDenominatorPostDelhi))
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
		Bor: &params.BorConfig{
			JaipurBlock: big.NewInt(23850000),
			Period: map[string]uint64{
				"0": 2,
			},
			ProducerDelay: map[string]uint64{
				"0": 6,
			},
			Sprint: map[string]uint64{
				"0": 64,
			},
			BackupMultiplier: map[string]uint64{
				"0": 2,
			},
			ValidatorContract:     "0x0000000000000000000000000000000000001000",
			StateReceiverContract: "0x0000000000000000000000000000000000001001",
			OverrideStateSyncRecords: map[string]int{
				"14949120": 8,
				"14949184": 0,
				"14953472": 0,
				"14953536": 5,
				"14953600": 0,
				"14953664": 0,
				"14953728": 0,
				"14953792": 0,
				"14953856": 0,
			},
			BurntContract: map[string]string{
				"23850000": "0x70bca57f4579f58670ab2d18ef16e02c17553c38",
			},
			BlockAlloc: map[string]interface{}{
				// write as interface since that is how it is decoded in genesis
				"22156660": map[string]interface{}{
					"0000000000000000000000000000000000001010": map[string]interface{}{
						"balance": "0x0",
						"code":    "0x60806040526004361061019c5760003560e01c806377d32e94116100ec578063acd06cb31161008a578063e306f77911610064578063e306f77914610a7b578063e614d0d614610aa6578063f2fde38b14610ad1578063fc0c546a14610b225761019c565b8063acd06cb31461097a578063b789543c146109cd578063cc79f97b14610a505761019c565b80639025e64c116100c65780639025e64c146107c957806395d89b4114610859578063a9059cbb146108e9578063abceeba21461094f5761019c565b806377d32e94146106315780638da5cb5b146107435780638f32d59b1461079a5761019c565b806347e7ef24116101595780637019d41a116101335780637019d41a1461053357806370a082311461058a578063715018a6146105ef578063771282f6146106065761019c565b806347e7ef2414610410578063485cc9551461046b57806360f96a8f146104dc5761019c565b806306fdde03146101a15780631499c5921461023157806318160ddd1461028257806319d27d9c146102ad5780632e1a7d4d146103b1578063313ce567146103df575b600080fd5b3480156101ad57600080fd5b506101b6610b79565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156101f65780820151818401526020810190506101db565b50505050905090810190601f1680156102235780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561023d57600080fd5b506102806004803603602081101561025457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610bb6565b005b34801561028e57600080fd5b50610297610c24565b6040518082815260200191505060405180910390f35b3480156102b957600080fd5b5061036f600480360360a08110156102d057600080fd5b81019080803590602001906401000000008111156102ed57600080fd5b8201836020820111156102ff57600080fd5b8035906020019184600183028401116401000000008311171561032157600080fd5b9091929391929390803590602001909291908035906020019092919080359060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610c3a565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6103dd600480360360208110156103c757600080fd5b8101908080359060200190929190505050610caa565b005b3480156103eb57600080fd5b506103f4610dfc565b604051808260ff1660ff16815260200191505060405180910390f35b34801561041c57600080fd5b506104696004803603604081101561043357600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610e05565b005b34801561047757600080fd5b506104da6004803603604081101561048e57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610fc1565b005b3480156104e857600080fd5b506104f1611090565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561053f57600080fd5b506105486110b6565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561059657600080fd5b506105d9600480360360208110156105ad57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506110dc565b6040518082815260200191505060405180910390f35b3480156105fb57600080fd5b506106046110fd565b005b34801561061257600080fd5b5061061b6111cd565b6040518082815260200191505060405180910390f35b34801561063d57600080fd5b506107016004803603604081101561065457600080fd5b81019080803590602001909291908035906020019064010000000081111561067b57600080fd5b82018360208201111561068d57600080fd5b803590602001918460018302840111640100000000831117156106af57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506111d3565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561074f57600080fd5b50610758611358565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156107a657600080fd5b506107af611381565b604051808215151515815260200191505060405180910390f35b3480156107d557600080fd5b506107de6113d8565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561081e578082015181840152602081019050610803565b50505050905090810190601f16801561084b5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561086557600080fd5b5061086e611411565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156108ae578082015181840152602081019050610893565b50505050905090810190601f1680156108db5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b610935600480360360408110156108ff57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919050505061144e565b604051808215151515815260200191505060405180910390f35b34801561095b57600080fd5b50610964611474565b6040518082815260200191505060405180910390f35b34801561098657600080fd5b506109b36004803603602081101561099d57600080fd5b8101908080359060200190929190505050611501565b604051808215151515815260200191505060405180910390f35b3480156109d957600080fd5b50610a3a600480360360808110156109f057600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019092919080359060200190929190505050611521565b6040518082815260200191505060405180910390f35b348015610a5c57600080fd5b50610a65611541565b6040518082815260200191505060405180910390f35b348015610a8757600080fd5b50610a90611546565b6040518082815260200191505060405180910390f35b348015610ab257600080fd5b50610abb61154c565b6040518082815260200191505060405180910390f35b348015610add57600080fd5b50610b2060048036036020811015610af457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506115d9565b005b348015610b2e57600080fd5b50610b376115f6565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b60606040518060400160405280600b81526020017f4d6174696320546f6b656e000000000000000000000000000000000000000000815250905090565b6040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260108152602001807f44697361626c656420666561747572650000000000000000000000000000000081525060200191505060405180910390fd5b6000601260ff16600a0a6402540be40002905090565b60006040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260108152602001807f44697361626c656420666561747572650000000000000000000000000000000081525060200191505060405180910390fd5b60003390506000610cba826110dc565b9050610cd18360065461161c90919063ffffffff16565b600681905550600083118015610ce657508234145b610d58576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f496e73756666696369656e7420616d6f756e740000000000000000000000000081525060200191505060405180910390fd5b8173ffffffffffffffffffffffffffffffffffffffff16600260009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167febff2602b3f468259e1e99f613fed6691f3a6526effe6ef3e768ba7ae7a36c4f8584610dd4876110dc565b60405180848152602001838152602001828152602001935050505060405180910390a3505050565b60006012905090565b610e0d611381565b610e1657600080fd5b600081118015610e535750600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1614155b610ea8576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526023815260200180611da76023913960400191505060405180910390fd5b6000610eb3836110dc565b905060008390508073ffffffffffffffffffffffffffffffffffffffff166108fc849081150290604051600060405180830381858888f19350505050158015610f00573d6000803e3d6000fd5b50610f168360065461163c90919063ffffffff16565b6006819055508373ffffffffffffffffffffffffffffffffffffffff16600260009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167f4e2ca0515ed1aef1395f66b5303bb5d6f1bf9d61a353fa53f73f8ac9973fa9f68585610f98896110dc565b60405180848152602001838152602001828152602001935050505060405180910390a350505050565b600760009054906101000a900460ff1615611027576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526023815260200180611d846023913960400191505060405180910390fd5b6001600760006101000a81548160ff02191690831515021790555080600260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555061108c8261165b565b5050565b600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60008173ffffffffffffffffffffffffffffffffffffffff16319050919050565b611105611381565b61110e57600080fd5b600073ffffffffffffffffffffffffffffffffffffffff166000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a360008060006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550565b60065481565b60008060008060418551146111ee5760009350505050611352565b602085015192506040850151915060ff6041860151169050601b8160ff16101561121957601b810190505b601b8160ff16141580156112315750601c8160ff1614155b156112425760009350505050611352565b60018682858560405160008152602001604052604051808581526020018460ff1660ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa15801561129f573d6000803e3d6000fd5b505050602060405103519350600073ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff16141561134e576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260128152602001807f4572726f7220696e2065637265636f766572000000000000000000000000000081525060200191505060405180910390fd5b5050505b92915050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614905090565b6040518060400160405280600181526020017f890000000000000000000000000000000000000000000000000000000000000081525081565b60606040518060400160405280600581526020017f4d41544943000000000000000000000000000000000000000000000000000000815250905090565b6000813414611460576000905061146e565b61146b338484611753565b90505b92915050565b6040518060800160405280605b8152602001611e1c605b91396040516020018082805190602001908083835b602083106114c357805182526020820191506020810190506020830392506114a0565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040516020818303038152906040528051906020012081565b60056020528060005260406000206000915054906101000a900460ff1681565b600061153761153286868686611b10565b611be6565b9050949350505050565b608981565b60015481565b604051806080016040528060528152602001611dca605291396040516020018082805190602001908083835b6020831061159b5780518252602082019150602081019050602083039250611578565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040516020818303038152906040528051906020012081565b6115e1611381565b6115ea57600080fd5b6115f38161165b565b50565b600260009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60008282111561162b57600080fd5b600082840390508091505092915050565b60008082840190508381101561165157600080fd5b8091505092915050565b600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16141561169557600080fd5b8073ffffffffffffffffffffffffffffffffffffffff166000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a3806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050565b6000803073ffffffffffffffffffffffffffffffffffffffff166370a08231866040518263ffffffff1660e01b8152600401808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060206040518083038186803b1580156117d357600080fd5b505afa1580156117e7573d6000803e3d6000fd5b505050506040513d60208110156117fd57600080fd5b8101908080519060200190929190505050905060003073ffffffffffffffffffffffffffffffffffffffff166370a08231866040518263ffffffff1660e01b8152600401808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060206040518083038186803b15801561188f57600080fd5b505afa1580156118a3573d6000803e3d6000fd5b505050506040513d60208110156118b957600080fd5b810190808051906020019092919050505090506118d7868686611c30565b8473ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff16600260009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167fe6497e3ee548a3372136af2fcb0696db31fc6cf20260707645068bd3fe97f3c48786863073ffffffffffffffffffffffffffffffffffffffff166370a082318e6040518263ffffffff1660e01b8152600401808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060206040518083038186803b1580156119df57600080fd5b505afa1580156119f3573d6000803e3d6000fd5b505050506040513d6020811015611a0957600080fd5b81019080805190602001909291905050503073ffffffffffffffffffffffffffffffffffffffff166370a082318e6040518263ffffffff1660e01b8152600401808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060206040518083038186803b158015611a9757600080fd5b505afa158015611aab573d6000803e3d6000fd5b505050506040513d6020811015611ac157600080fd5b8101908080519060200190929190505050604051808681526020018581526020018481526020018381526020018281526020019550505050505060405180910390a46001925050509392505050565b6000806040518060800160405280605b8152602001611e1c605b91396040516020018082805190602001908083835b60208310611b625780518252602082019150602081019050602083039250611b3f565b6001836020036101000a03801982511681845116808217855250505050505090500191505060405160208183030381529060405280519060200120905060405181815273ffffffffffffffffffffffffffffffffffffffff8716602082015285604082015284606082015283608082015260a0812092505081915050949350505050565b60008060015490506040517f190100000000000000000000000000000000000000000000000000000000000081528160028201528360228201526042812092505081915050919050565b3073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff161415611cd2576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f63616e27742073656e6420746f204d524332300000000000000000000000000081525060200191505060405180910390fd5b8173ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015611d18573d6000803e3d6000fd5b508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef836040518082815260200191505060405180910390a350505056fe54686520636f6e747261637420697320616c726561647920696e697469616c697a6564496e73756666696369656e7420616d6f756e74206f7220696e76616c69642075736572454950373132446f6d61696e28737472696e67206e616d652c737472696e672076657273696f6e2c75696e7432353620636861696e49642c6164647265737320766572696679696e67436f6e747261637429546f6b656e5472616e736665724f726465722861646472657373207370656e6465722c75696e7432353620746f6b656e49644f72416d6f756e742c6279746573333220646174612c75696e743235362065787069726174696f6e29a265627a7a72315820a4a6f71a98ac3fc613c3a8f1e2e11b9eb9b6b39f125f7d9508916c2b8fb02c7164736f6c63430005100032",
					},
				},
			},
		},
	}

	ec, err := ethclient.Dial("https://polygon-rpc.com")
	require.NoError(t, err)
	defer ec.Close()

	userDir, _ := os.UserHomeDir()
	f, err := os.Create(filepath.Join(userDir, "work", "base-fee-calc.tsv"))
	require.NoError(t, err)
	defer f.Close()

	startBlock := int64(35893346)
	var newFeeCalc *big.Int
	for i := startBlock; i < startBlock+int64(24*60*30); i++ {
		blk, err := ec.BlockByNumber(context.Background(), big.NewInt(i))
		if err != nil {
			continue
		}

		h := blk.Header()
		// checkFeeCalc := misc.CalcBaseFee(BorMainnetChainConfig, h)
		checkFeeCalc := h.BaseFee

		if newFeeCalc != nil {
			h.BaseFee = newFeeCalc
		}

		BorMainnetChainConfig.Bor.DelhiBlock = big.NewInt(i)
		newFeeCalc = misc.CalcBaseFee(BorMainnetChainConfig, h)
		BorMainnetChainConfig.Bor.DelhiBlock = nil
		// newFeeCalc = NewCalcBaseFee(BorMainnetChainConfig, h)

		gasPercent := h.GasUsed * 100 / h.GasLimit

		s := fmt.Sprintf("%v\t%.1f\t%.1f\t%v\n", i,
			float64(checkFeeCalc.Int64())/1_000_000_000,
			float64(newFeeCalc.Int64())/1_000_000_000,
			gasPercent)

		f.Write([]byte(s))
	}
}
