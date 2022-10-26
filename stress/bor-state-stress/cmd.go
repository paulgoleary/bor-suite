package main

import (
	"crypto/rand"
	"fmt"
	"github.com/maticnetwork/bor/core/rawdb"
	"github.com/maticnetwork/bor/ethdb"
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

	dbPath := argsOnly[0]

	var sourceDb ethdb.Database
	var err error
	cntFound := 0
	if sourceDb, err = rawdb.NewLevelDBDatabaseWithFreezer(dbPath, cacheMb, 1024, filepath.Join(dbPath, "ancient"), ""); err == nil {
		defer sourceDb.Close()
		startTotal := time.Now()
		for x := 0; x < 10; x++ {
			startBatch := time.Now()
			for i := 0; i < 100_000; i++ {
				var randKey [32]byte
				rand.Read(randKey[:])
				if bytes, _ := sourceDb.Get(randKey[:]); bytes != nil {
					cntFound++
				}
			}
			durationBatch := time.Since(startBatch)
			fmt.Printf("BATCH %v execution time: %v\n", x, durationBatch)
		}
		durationTotal := time.Since(startTotal)
		fmt.Printf("TOTAL execution time: %v\n", durationTotal)
	}
}
