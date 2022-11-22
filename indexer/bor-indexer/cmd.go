package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/gammazero/workerpool"
	"github.com/paulgoleary/bor-suite/indexer"
	"os"
	"path/filepath"
	"strconv"
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
	argsOnly := os.Args[1:]
	cacheMb := 512
	if len(argsOnly) > 1 {
		if mb, err := strconv.Atoi(argsOnly[1]); err != nil {
			panic(err)
		} else {
			cacheMb = mb
		}
	}

	borDbPath := argsOnly[0]

	sink := testSink{indexer.MakeEventCounterSink("Transfer")}

	var sourceDb ethdb.Database
	var err error
	if sourceDb, err = rawdb.NewLevelDBDatabaseWithFreezer(borDbPath, cacheMb, 1024, filepath.Join(borDbPath, "ancient"), "", true); err == nil {
		defer sourceDb.Close()

		const numWorkers = 8
		wp := workerpool.New(numWorkers)

		startTotal := time.Now()

		sliceFunc := func(blockMod uint64) {
			cntReceipts := 0
			cntProc := 0
			blockNum := blockMod
			for {
				blockHash := rawdb.ReadCanonicalHash(sourceDb, blockNum)
				if theBlock := rawdb.ReadBlock(sourceDb, blockHash, blockNum); theBlock == nil {
					break
				} else {
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
