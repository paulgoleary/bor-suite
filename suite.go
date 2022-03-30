package bor_suite

import (
	"github.com/stretchr/testify/suite"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/node"
)

type BorServerTestSuite struct {
	suite.Suite

	testConfig node.Config

	testNode *node.Node
}

func getRandomNetworkID(min, max int) uint64 {
	rand.Seed(time.Now().UnixNano())
	return uint64(rand.Intn(max-min) + min)
}

func (s *BorServerTestSuite) SetupSuite() {

	nwid := getRandomNetworkID(80000, 90000)
	s.initNode(nwid)

}

var ethInWei *big.Int

func init() {
	ethInWei = new(big.Int)
	ethInWei.SetString("1000000000000000000", 10)

}

func (s *BorServerTestSuite) initNode(nwid uint64) {

	genesisConfig := GenesisConfig{}
	genesisConfig.EndowAmount = new(big.Int).Mul(ethInWei, big.NewInt(100))
	genesisConfig.ValidatorsAmount = new(big.Int).Mul(ethInWei, big.NewInt(100))
	genesisConfig.ChainId = big.NewInt(0).SetInt64(int64(nwid))
	genesisConfig.Validators = []string{devValidator}
	genesisConfig.Gaslimit = 20000000

	s.testConfig.HTTPModules = append(s.testConfig.HTTPModules, []string{"eth", "admin", "txpool", "clique"}...)

	err := initNode(&genesisConfig, &s.testConfig, true)
	s.NoError(err)

	if _, err = startDevServer(&s.testConfig, nwid); err != nil {
		panic(err)
	}

}
