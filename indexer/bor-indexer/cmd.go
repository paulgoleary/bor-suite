package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
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
		startTotal := time.Now()

		blockNum := uint64(0)
		cntReceipts := 0
		for {
			blockHash := rawdb.ReadCanonicalHash(sourceDb, blockNum)
			if theBlock := rawdb.ReadBlock(sourceDb, blockHash, blockNum); theBlock == nil {
				break
			} else {
				if rawdb.ReadRawReceipts(sourceDb, blockHash, blockNum) != nil {
					cntReceipts++
				}
				blockNum++
				if blockNum%100_000 == 0 {
					fmt.Printf("AT BLOCK %v, receipt count %v\n", blockNum, cntReceipts)
				}
			}
		}

		durationTotal := time.Since(startTotal)
		fmt.Printf("TOTAL execution time: %v\n", durationTotal)
	} else {
		fmt.Printf("ERROR: %v\n", err)
	}
}
