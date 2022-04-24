package state_turbo

import (
	"context"
	"github.com/holiman/uint256"
	ecommon "github.com/ledgerwatch/erigon/common"

	estate "github.com/ledgerwatch/erigon/core/state"
	"github.com/ledgerwatch/erigon/ethdb"

	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestErigonStateBasics(t *testing.T) {

	dir, err := ioutil.TempDir("", "test-erigon-state-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	testDest := ethdb.NewMDBX().Path(dir).MustOpen()
	// testDest := ethdb.NewLMDB().Path(dir).MustOpen()
	testRwTx, err := testDest.BeginRw(context.Background())
	require.NoError(t, err)

	contractAddr := ecommon.HexToAddress("0xDEADBEEF")

	r, tsw := estate.NewPlainStateReader(testRwTx), estate.NewPlainStateWriter(testRwTx, nil, 0)
	intraBlockState := estate.New(r)
	intraBlockState.CreateAccount(ecommon.Address(contractAddr), true)
	_ = tsw

	testKey1 := ecommon.HexToHash("0xCAFEBABE1")
	testKey2 := ecommon.HexToHash("0xCAFEBABE2")
	intraBlockState.SetState(contractAddr, &testKey1, *uint256.NewInt(1))
	intraBlockState.SetState(contractAddr, &testKey2, *uint256.NewInt(2))

	var checkState uint256.Int
	intraBlockState.GetState(contractAddr, &testKey2, &checkState)
	require.Equal(t, uint64(2), checkState.Uint64())

}
