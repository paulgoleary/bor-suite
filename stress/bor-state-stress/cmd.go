package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"os"
	"path/filepath"
	"time"
)

func main() {

	var dbPath string
	var cacheMb, numBatches, numIterations int
	var err error

	if dbPath, cacheMb, numBatches, numIterations, err = processArgs(); err != nil {
		flag.Usage()
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	var sourceDb ethdb.Database
	cntFound := 0
	if sourceDb, err = rawdb.NewLevelDBDatabaseWithFreezer(dbPath, cacheMb, 1024, filepath.Join(dbPath, "ancient"), "", true); err == nil {
		defer sourceDb.Close()
		startTotal := time.Now()
		for x := 0; x < numBatches; x++ {
			startBatch := time.Now()
			for i := 0; i < numIterations; i++ {
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
	} else {
		fmt.Printf("ERROR: %v\n", err)
	}
}

func processArgs() (dbPath string, cacheMb, numBatches, numIterations int, err error) {
	flag.IntVar(&cacheMb, "cache", 512, "cache size in MB")
	flag.IntVar(&numBatches, "batches", 10, "number of batches to run")
	flag.IntVar(&numIterations, "iterations", 100_000, "number of iterations per batch")

	flag.StringVar(&dbPath, "path", "", "path to database")
	flag.Parse()
	if _, err = os.Stat(dbPath); err != nil {
		return
	}
	return
}
