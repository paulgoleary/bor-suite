package bor_suite

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func readGenesisEthConfig() (cfg *eth.Config, err error) {
	var genesisData []byte
	if genesisData, err = ioutil.ReadFile("./testdata/genesis.json"); err != nil {
		return
	}
	gen := &core.Genesis{}
	if err = json.Unmarshal(genesisData, gen); err != nil {
		return
	}

	cfg = &eth.Config{
		Genesis: gen,
	}

	return
}

func TestSetup(t *testing.T) {

	testDb := rawdb.NewMemoryDatabase()

	ethConf, err := readGenesisEthConfig()
	require.NoError(t, err)

	ethConf.Genesis.MustCommit(testDb)
	//ethereum := utils.CreateBorEthereum(ethConf)
	//ethConf.Genesis.MustCommit(ethereum.ChainDb())
}
