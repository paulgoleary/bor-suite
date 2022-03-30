package bor_suite

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/node"
	"math/big"
	"strings"
)

type GenesisConfig struct {
	EndowAmount      *big.Int
	Validators       []string
	ValidatorsAmount *big.Int
	Gaslimit         uint64
	ChainId          *big.Int
	BlockTime        uint64
	Epoch            uint64
}

// from clique.go
const (
	CliqueExtraVanity          = 32 // Fixed number of extra-data prefix bytes reserved for signer vanity
	CliqueExtraSeal            = 65 // Fixed number of extra-data suffix bytes reserved for signer seal
	errMissingVanity           = "extra-data 32 byte vanity prefix missing"
	errMissingSignature        = "extra-data 65 byte signature suffix missing"
	errInvalidExtraDataSigners = "invalid signer list in extra data"
)

// TODO: change !!!
const defaultGenesis = `
{
"config": {
"homesteadBlock": 0,
"chainId": 955301,
"eip150Block": 0,
"eip155Block": 0,
"eip158Block": 0,
"eip160Block": 0,
"ByzantiumBlock": 0,
"clique": {
"period": 0,
"epoch": 1000
}
},
"nonce": "0x0",
"timestamp": "0x00",
"extraData": "",
"gasLimit": "0x1312D00",
"difficulty": "0x1",
"mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
"coinbase": "0x0000000000000000000000000000000000000000",
"parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
"alloc": {}
}
`

func loadDefaultGenesis() (gen core.Genesis, err error) {
	if err = json.NewDecoder(strings.NewReader(defaultGenesis)).Decode(&gen); err != nil {
		return
	}
	return
}

func createGenesis(gc *GenesisConfig, nodeConfig *node.Config) (gen core.Genesis, err error) {

	if gen, err = loadDefaultGenesis(); err != nil {
		return
	}

	var signers []common.Address
	for i := range gc.Validators {
		signerAddr := common.HexToAddress(gc.Validators[i])
		signers = append(signers, signerAddr)
	}

	gen.Config.Clique.Epoch = gc.Epoch
	gen.Config.Clique.Period = gc.BlockTime

	ks := keystore.NewKeyStore(nodeConfig.KeyStoreDir, keystore.LightScryptN, keystore.LightScryptP)

	gen.Alloc = core.GenesisAlloc{}
	gen.ExtraData = make([]byte, CliqueExtraVanity+common.AddressLength*len(signers)+CliqueExtraSeal)
	for index, signerAddr := range signers {
		gen.Alloc[signerAddr] = core.GenesisAccount{Nonce: 0, Balance: gc.ValidatorsAmount}
		copy(gen.ExtraData[CliqueExtraVanity+index*common.AddressLength:], signerAddr[:])
	}

	isAccountSigner := func(account string) bool {
		for _, v := range signers {
			if strings.ToLower(account) == strings.ToLower(v.String()) {
				return true
			}
		}
		return false
	}

	for _, vi := range ks.Accounts() {
		acc := vi.Address.String()
		isSigner := isAccountSigner(acc)
		if !isSigner {
			gen.Alloc[vi.Address] = core.GenesisAccount{Nonce: 0, Balance: gc.EndowAmount}
		}
	}

	gen.GasLimit = gc.Gaslimit
	gen.Config.ChainID = gc.ChainId

	return
}
