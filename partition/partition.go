package main

import (
	"fmt"
	"github.com/gammazero/workerpool"
	"github.com/maticnetwork/bor/core/rawdb"
	"github.com/maticnetwork/bor/ethdb"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	argsWithoutProg := os.Args[1:]

	dbPath := argsWithoutProg[0]

	if sourceDb, err := rawdb.NewLevelDBDatabaseWithFreezer(dbPath, 0, 0, filepath.Join(dbPath, "ancient"), ""); err != nil {
		panic(err)
	} else {

		defer sourceDb.Close()

		db0Path := dbPath + ".0"
		db1Path := dbPath + ".1"

		var part0Db, part1Db ethdb.Database

		if part0Db, err = rawdb.NewLevelDBDatabaseWithFreezer(db0Path, 0, 0, filepath.Join(db0Path, "ancient"), ""); err != nil {
			panic(err)
		}
		defer part0Db.Close()

		if part1Db, err = rawdb.NewLevelDBDatabaseWithFreezer(db1Path, 0, 0, filepath.Join(db1Path, "ancient"), ""); err != nil {
			panic(err)
		}
		defer part1Db.Close()

		wp := workerpool.New(runtime.NumCPU() * 2)

		for i := 0; i < 0x100; i++ {
			partPrefix := byte(i)
			wp.Submit(func() {
				sliceTarget := part0Db
				if partPrefix >= 0x80 {
					sliceTarget = part1Db
				}
				iterSlice := sourceDb.NewIterator([]byte{partPrefix}, nil)
				currBatch := sliceTarget.NewBatch()
				iterCnt := 0
				for iterSlice.Next() {
					if sliceErr := currBatch.Put(iterSlice.Key(), iterSlice.Value()); sliceErr != nil {
						panic(sliceErr)
					} else if currBatch.ValueSize() > ethdb.IdealBatchSize {
						if sliceErr = currBatch.Write(); sliceErr != nil {
							panic(err)
						}
						currBatch.Reset()
					}
					iterCnt++
					if iterCnt%10000 == 0 {
						subPrefix := 0
						if len(iterSlice.Key()) > 1 {
							subPrefix = int(iterSlice.Key()[1])
						}
						println(fmt.Sprintf("slice prefix '%v.%v', iteration count %v", partPrefix, subPrefix, iterCnt))
					}
				}
				println(fmt.Sprintf("FINISHED slice prefix '%v', iteration count %v", partPrefix, iterCnt))
			})
		}
		wp.StopWait()
	}
}
