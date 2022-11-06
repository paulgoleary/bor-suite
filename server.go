package bor_suite

import (
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/node"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

// TODO: change !!!
const devValidator = "0x81679e1b01aa38c04ca5aec757432d10fda01dfd"

func initNode(gc *GenesisConfig, nodeConfig *node.Config, overwrite bool) error {

	if genesis, err := createGenesis(gc, nodeConfig); err != nil {
		return err
	} else {
		if stack, err := node.New(nodeConfig); err != nil {
			return err
		} else {
			if err = initDatabases(stack, &genesis, nodeConfig, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func initDatabases(stack *node.Node, genesis *core.Genesis, cfg *node.Config, overwrite bool) error {

	if overwrite {
		path := filepath.Join(cfg.DataDir, cfg.Name)
		if err := os.RemoveAll(cfg.ResolvePath(path)); err != nil {
			return err
		}
	}

	const chainDataPath = "chaindata"
	if _, err := os.Stat(cfg.ResolvePath(chainDataPath)); !os.IsNotExist(err) {
		if !overwrite {
			return fmt.Errorf("should not happen - databases exist and overwrite was not specified or failed")
		}
	}

	if chaindb, err := stack.OpenDatabase(chainDataPath, 0, 0, "", false); err != nil {
		return err
	} else {
		defer chaindb.Close()

		if _, _, err = core.SetupGenesisBlock(chaindb, genesis); err != nil {
			return err
		}
	}

	return nil
}

func startDevServer(nodeConfig *node.Config, nwid uint64) (n *node.Node, err error) {

	if n, err = node.New(nodeConfig); err != nil {
		return
	}

	ethConfig := ethconfig.Defaults
	ethConfig.NetworkId = nwid
	ethConfig.SyncMode = downloader.FullSync
	ethConfig.Miner.GasFloor = 20_000_000
	ethConfig.Miner.GasCeil = 20_000_000

	utils.RegisterEthService(n, &ethConfig)

	app := &cli.App{}
	set := flag.NewFlagSet("dev", 0)
	cxt := cli.NewContext(app, set, nil)
	utils.StartNode(cxt, n, false)

	am := n.AccountManager()
	_ = am

	return
}
