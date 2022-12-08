package txpool

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"math/big"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

// http://35.172.112.23:8545 ??
func getEthClient() (*jsonrpc.Client, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "polygon-rpc.com",
	}
	return jsonrpc.NewClient(u.String())
}

func TestBasicContent(t *testing.T) {

	t.SkipNow() //

	ec, err := getEthClient()
	require.NoError(t, err)
	defer ec.Close()

	var blockMap map[string]interface{}
	err = ec.Call("eth_getBlockByNumber", &blockMap, ethgo.Latest.String(), false)
	require.NoError(t, err)

	getMapNumber := func(m map[string]interface{}, n string) *big.Int {
		numStr := m[n].(string)
		num := new(big.Int)
		var ok bool
		numStr = strings.TrimPrefix(numStr, "0x")
		num, ok = num.SetString(numStr, 16)
		require.True(t, ok)
		return num
	}

	baseFee := getMapNumber(blockMap, "baseFeePerGas")

	checkGasPrice, err := ec.Eth().GasPrice()
	require.NoError(t, err)
	require.True(t, checkGasPrice > 0)

	fmt.Printf("Block number: %v, current GAS PRICE: %v, base fee: %v", getMapNumber(blockMap, "number"), checkGasPrice/1_000_000_000, baseFee)

	var out map[string]interface{}
	err = ec.Call("txpool_content", &out)
	require.NoError(t, err)

	txPoolBytes, err := json.Marshal(&out)
	require.NoError(t, err)

	err = os.WriteFile("/Users/pauloleary/work/bor-txpool-20221027-1.json", txPoolBytes, 0644)
	require.NoError(t, err)
}

func TestTxPoolProcessing(t *testing.T) {

	t.SkipNow() // todo: fix this test

	txPoolBytes, err := os.ReadFile("/Users/pauloleary/work/bor-txpool-20221027-1.json")
	require.NoError(t, err)

	var mapPool map[string]map[common.Address]map[string]*types.Transaction
	err = json.Unmarshal(txPoolBytes, &mapPool)
	require.NoError(t, err)

	mapPending := mapPool["pending"]
	fmt.Printf("PENDING transactions: %v\n", len(mapPending))

	pending := make(map[common.Address]types.Transactions)
	for k, v := range mapPending {
		var tt types.Transactions
		for _, tv := range v {
			tt = append(tt, tv)
		}
		pending[k] = tt
	}

	// fake signer - Bor
	signer := types.NewEIP155Signer(big.NewInt(137))

	baseFee := big.NewInt(30)
	baseFee = baseFee.Mul(baseFee, big.NewInt(1000000000))

	info := CheckTransactionsByPriceAndNonce(signer, pending, baseFee)
	fmt.Printf("Included: %v, too low: %v, bad from: %v\n", info.included, info.tooLow, info.badFrom)

	start := time.Now()
	cntBefore := len(pending)
	txs := types.NewTransactionsByPriceAndNonce(signer, pending, baseFee)
	require.True(t, txs.Peek() != nil)
	cntAfter := len(pending)
	elapsed := time.Since(start)
	fmt.Printf("Tx processing took: %v, count before, after: %v, %v\n", elapsed, cntBefore, cntAfter)
}

type checkTxInfo struct {
	included int
	tooLow   int
	badFrom  int

	feeCapHigh int
	tipCapHigh int

	feeCapHist [1000]int
	tipCapHist [1000]int
}

var gwei = big.NewInt(1_000_000_000)

func CheckTransactionsByPriceAndNonce(signer types.Signer, txs map[common.Address]types.Transactions, baseFee *big.Int) (info checkTxInfo) {
	for from, accTxs := range txs {
		acc, _ := types.Sender(signer, accTxs[0])
		if _, err := types.NewTxWithMinerFee(accTxs[0], baseFee); err != nil {
			info.tooLow++
			continue
		} else if acc != from {
			info.badFrom++
			continue
		}

		// gasFeeCap.Cmp(st.gasTipCap)
		if accTxs[0].GasFeeCap().Cmp(accTxs[0].GasTipCap()) < 0 {
			info.tooLow++
			continue
		}

		feeCapGwei := new(big.Int).Div(accTxs[0].GasFeeCap(), gwei).Int64()
		if feeCapGwei >= int64(len(info.feeCapHist)) {
			info.feeCapHigh++
		} else {
			info.feeCapHist[feeCapGwei] += 1
		}

		tipCapGwei := new(big.Int).Div(accTxs[0].GasTipCap(), gwei).Int64()
		if tipCapGwei >= int64(len(info.tipCapHist)) {
			info.tipCapHigh++
		} else {
			info.tipCapHist[tipCapGwei] += 1
		}

		info.included++
	}
	return
}

func TestStuckTransaction(t *testing.T) {

	t.SkipNow() // todo: fix this test

	txJson := `{
         "blockHash":null,
         "blockNumber":null,
         "from":"0x1707128c8ab9829d5c632957da6e4cd6732d46b1",
         "gas":"0x895440",
         "gasPrice":"0x895440",
         "hash":"0x1a824ec1e16cf03b196cf65a482d3a449b8e8002de900491f6177d5919c8bd6b",
         "input":"0x38ed17390000000000000000000000000000000000000000000000000de0b6b3a7640000000000000000000000000000000000000000000000000000000000000000271000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000001707128c8ab9829d5c632957da6e4cd6732d46b1000000000000000000000000000000000000000000000000000000006280440b00000000000000000000000000000000000000000000000000000000000000030000000000000000000000008f3cf7ad23cd3cadbd9735aff958023239c6a063000000000000000000000000831753dd7087cac61ab5644b308642cc1c33dc130000000000000000000000002791bca1f2de4661ed88a30c99a7a9449aa84174",
         "nonce":"0x654",
         "to":"0xa5e0829caced8ffdd4de3c43696c57f7d7a678ff",
         "transactionIndex":null,
         "value":"0x0",
         "type":"0x0",
         "v":"0x136",
         "r":"0x8bf77dcb1cad161d5b12a2ef27bdea3ca2c31dcc643bc14b979c8ce056cdcb99",
         "s":"0xe3cc1b7dae49f5869c0beab447ea8080073a6d716202c6d266dc1fa783eaebd"
      }`

	var tx types.Transaction
	err := json.Unmarshal([]byte(txJson), &tx)
	require.NoError(t, err)

	signer := types.NewEIP155Signer(big.NewInt(137))
	acc, _ := types.Sender(signer, &tx)
	require.Equal(t, common.HexToAddress("0x1707128c8ab9829d5c632957da6e4cd6732d46b1"), acc)

	baseFee := big.NewInt(1)
	baseFee = baseFee.Mul(baseFee, big.NewInt(1000000000))

	_, err = types.NewTxWithMinerFee(&tx, baseFee)
	require.NoError(t, err)
}
