package state_bor

import (
	"github.com/maticnetwork/bor/common"
	"github.com/maticnetwork/bor/core/rawdb"
	"github.com/maticnetwork/bor/core/state"
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

	testState, err := state.New(common.Hash{}, state.NewDatabase(testDb), nil)
	require.NoError(t, err)

	testAcct := common.HexToAddress("0xDEADBEEF")
	testState.CreateAccount(testAcct)

	testState.SetState(testAcct, common.HexToHash("0xCAFEBABE1"), common.HexToHash("0xCAFEBEEF1"))
	testState.SetState(testAcct, common.HexToHash("0xCAFEBABE2"), common.HexToHash("0xCAFEBEEF2"))

	checkState1 := testState.GetState(testAcct, common.HexToHash("0xCAFEBABE1"))
	require.Equal(t, common.HexToHash("0xCAFEBEEF1").Bytes(), checkState1.Bytes())

	_, err = testState.Commit(true)
	require.NoError(t, err)
}
