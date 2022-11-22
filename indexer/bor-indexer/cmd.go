package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/gammazero/workerpool"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

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
			for {
				blockHash := rawdb.ReadCanonicalHash(sourceDb, blockMod)
				if theBlock := rawdb.ReadBlock(sourceDb, blockHash, blockMod); theBlock == nil {
					break
				} else {
					if rawdb.ReadRawReceipts(sourceDb, blockHash, blockMod) != nil {
						cntReceipts++
					}
					blockMod += numWorkers
					cntProc++
					if cntProc%100_000 == 0 {
						fmt.Printf("WORKER %v: AT BLOCK %v, receipt count %v\n", blockMod, cntProc, cntReceipts)
					}
				}
			}
		}

		for i := uint64(0); i < numWorkers; i++ {
			wp.Submit(func() {
				sliceFunc(i)
			})
		}

		wp.StopWait()

		durationTotal := time.Since(startTotal)
		fmt.Printf("TOTAL execution time: %v\n", durationTotal)
	} else {
		fmt.Printf("ERROR: %v\n", err)
	}
}
