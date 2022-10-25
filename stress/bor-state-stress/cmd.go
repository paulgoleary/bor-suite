package main

import (
	"crypto/rand"
	"fmt"
	"github.com/maticnetwork/bor/core/rawdb"
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
	if sourceDb, err := rawdb.NewLevelDBDatabaseWithFreezer(dbPath, cacheMb, 1024, filepath.Join(dbPath, "ancient"), ""); err == nil {
		defer sourceDb.Close()
		startTotal := time.Now()
		for x := 0; x < 10; x++ {
			startBatch := time.Now()
			for i := 0; i < 100_000; i++ {
				var randKey [20]byte
				rand.Read(randKey[:])
				iter := sourceDb.NewIterator(nil, randKey[:])
				iter.Next()
			}
			durationBatch := time.Since(startBatch)
			println(fmt.Sprintf("BATCH %v execution time: %v", x, durationBatch))
		}
		durationTotal := time.Since(startTotal)
		println(fmt.Sprintf("TOTAL execution time: %v", durationTotal))
	}
}
