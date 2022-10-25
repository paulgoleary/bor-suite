package state_bor

import (
	"github.com/maticnetwork/bor/common"
	"github.com/maticnetwork/bor/core/rawdb"
	"github.com/maticnetwork/bor/core/state"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestBorStateBasics(t *testing.T) {

	dir, err := ioutil.TempDir("", "test-bor-state-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	testDb, err := rawdb.NewLevelDBDatabase(dir, 0, 0, "")
	require.NoError(t, err)

	sdb := state.NewDatabase(testDb)
	testState, err := state.New(common.Hash{}, sdb, nil)
	require.NoError(t, err)

	testAcct := common.HexToAddress("DEADBEEF")
	testState.CreateAccount(testAcct)

	testState.SetState(testAcct, common.HexToHash("CAFEBABE01"), common.HexToHash("CAFEBEEF01"))
	testState.SetState(testAcct, common.HexToHash("CAFEBABE02"), common.HexToHash("CAFEBEEF02"))

	checkState1 := testState.GetState(testAcct, common.HexToHash("0xCAFEBABE01"))
	require.Equal(t, common.HexToHash("0xCAFEBEEF01").Bytes(), checkState1.Bytes())

	_, err = testState.Commit(false)
	require.NoError(t, err)
}

func TestDBBasics(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-bor-state-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	testDb, err := rawdb.NewLevelDBDatabase(dir, 0, 0, "")
	require.NoError(t, err)

	testDb.Put(common.Hex2Bytes("CAFEBABE"), common.Hex2Bytes("CAFEBEEF"))

	checkIter := testDb.NewIterator(nil, common.Hex2Bytes("CAFE"))
	for checkIter.Next() {
		println(hexutils.BytesToHex(checkIter.Key()))
	}
}
