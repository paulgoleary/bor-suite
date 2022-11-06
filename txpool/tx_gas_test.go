package txpool

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestTxGasHistogram(t *testing.T) {

	ec, err := ethclient.Dial("https://polygon-rpc.com")
	require.NoError(t, err)
	defer ec.Close()

	var txGasHisto [30]int
	var blkGasHisto [30]int
	startBlock := int64(35181438)
	errorCnt := 0
	for i := int64(0); i < 30*60*24; i++ {
		blk, err := ec.BlockByNumber(context.Background(), big.NewInt(startBlock+i))
		if err != nil {
			errorCnt++
			continue
		}
		if i%100 == 0 {
			fmt.Printf("tx gas histo: %v\n", txGasHisto)
			fmt.Printf("blk gas histo: %v\n", blkGasHisto)
		}

		blkGasHisto[blk.GasUsed()/1_000_000] += 1
		for _, tx := range blk.Transactions() {
			txGasHisto[tx.Gas()/1_000_000] += 1
		}
	}
	fmt.Printf("FINAL tx gas histo: %v\n", txGasHisto)
	fmt.Printf("FINAL blk gas histo: %v\n", blkGasHisto)
	fmt.Printf("total ERRORs: %v\n", errorCnt)
}
