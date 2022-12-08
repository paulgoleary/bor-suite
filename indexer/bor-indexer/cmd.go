package main

import (
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/gammazero/workerpool"
	"github.com/paulgoleary/bor-suite/indexer"
	"os"
	"path/filepath"
	"time"
)

type testSink struct {
	sink *indexer.EventCounterSink
}

func (t *testSink) ProcessBorLogs(logs []*types.Log) error {
	t.sink.Process(indexer.BorLogsToEthGo(logs))
	return nil
}

var _ indexer.BorLogSink = &testSink{}

func main() {

	var cacheMb int
	var borDbPath string
	var err error

	if borDbPath, cacheMb, err = processArgs(); err != nil {
		flag.Usage()
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	sink := testSink{indexer.MakeEventCounterSink("Transfer")}

	var sourceDb ethdb.Database
	if sourceDb, err = rawdb.NewLevelDBDatabaseWithFreezer(borDbPath, cacheMb, 1024, filepath.Join(borDbPath, "ancient"), "", true); err == nil {
		defer sourceDb.Close()

		const numWorkers = 8
		wp := workerpool.New(numWorkers)

		startTotal := time.Now()

		var nullHash common.Hash

		sliceFunc := func(blockMod uint64) {
			cntReceipts := 0
			cntProc := 0
			blockNum := blockMod
			for {
				var blockHash common.Hash
				if blockHash = rawdb.ReadCanonicalHash(sourceDb, blockNum); blockHash == nullHash {
					break
				}
				//if theBlock := rawdb.ReadBlock(sourceDb, blockHash, blockNum); theBlock == nil {
				//}
				if rcpts := rawdb.ReadRawReceipts(sourceDb, blockHash, blockNum); rcpts != nil {
					for i := range rcpts {
						sink.ProcessBorLogs(rcpts[i].Logs) // TODO: errors, etc.
					}
					cntReceipts++
				}
				blockNum += numWorkers
				cntProc++
				if cntProc%100_000 == 0 {
					fmt.Printf("WORKER %v: AT BLOCK %v, count processed, receipts %v, %v. FOUND %v\n", blockMod,
						blockNum, cntProc, cntReceipts, sink.sink.Count.Load())
				}
			}
		}

		for i := uint64(0); i < numWorkers; i++ {
			x := i
			f := func() {
				sliceFunc(x)
			}
			wp.Submit(f)
		}

		wp.StopWait()

		durationTotal := time.Since(startTotal)
		fmt.Printf("TOTAL execution time: %v\n", durationTotal)
		fmt.Printf("FOUND EVENTS: %v\n", sink.sink.Count.Load())
	} else {
		fmt.Printf("ERROR: %v\n", err)
	}
}

func processArgs() (dbPath string, cacheMb int, err error) {
	flag.IntVar(&cacheMb, "cache", 512, "cache size in MB")
	flag.StringVar(&dbPath, "path", "", "path to database")
	flag.Parse()
	if _, err = os.Stat(dbPath); err != nil {
		return
	}
	return
}
