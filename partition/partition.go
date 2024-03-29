package partition

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/gammazero/workerpool"
	"path/filepath"
	"runtime"
)

func POCPartitionDatabase(dbPath, targetPath string, numParts int, report func(string)) error {

	if numParts != 2 {
		return fmt.Errorf("should not happen - POC version only supports two partitions")
	}

	if sourceDb, err := rawdb.NewLevelDBDatabaseWithFreezer(dbPath, 1024, 1024, filepath.Join(dbPath, "ancient"), "", true); err != nil {
		panic(err)
	} else {

		defer sourceDb.Close()

		db0Path := filepath.Join(targetPath, ".0")
		db1Path := filepath.Join(targetPath, ".1")

		var part0Db, part1Db ethdb.Database

		if part0Db, err = rawdb.NewLevelDBDatabase(db0Path, 1024, 1024, "", false); err != nil {
			panic(err)
		}
		defer part0Db.Close()

		if part1Db, err = rawdb.NewLevelDBDatabase(db1Path, 1024, 1024, "", false); err != nil {
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
							panic(sliceErr)
						}
						currBatch.Reset()
					}
					iterCnt++
					if iterCnt%10000 == 0 {
						subPrefix := 0
						if len(iterSlice.Key()) > 1 {
							subPrefix = int(iterSlice.Key()[1])
						}
						report(fmt.Sprintf("slice prefix '%v.%v', iteration count %v", partPrefix, subPrefix, iterCnt))
					}
				}
				// make sure we write out the last batch
				if err = currBatch.Write(); err != nil {
					panic(err)
				}
				report(fmt.Sprintf("FINISHED slice prefix '%v', iteration count %v", partPrefix, iterCnt))
			})
		}
		wp.StopWait()
	}

	return nil
}

func POCPartitionDatabase2(dbPath, targetPath string, numParts int, report func(string)) error {

	if numParts != 2 {
		return fmt.Errorf("should not happen - POC version only supports two partitions")
	}

	if sourceDb, err := rawdb.NewLevelDBDatabaseWithFreezer(dbPath, 2024, 1024, filepath.Join(dbPath, "ancient"), "", true); err != nil {
		panic(err)
	} else {

		defer sourceDb.Close()

		db0Path := filepath.Join(targetPath, ".0")
		db1Path := filepath.Join(targetPath, ".1")

		var part0Db, part1Db ethdb.Database

		if part0Db, err = rawdb.NewLevelDBDatabase(db0Path, 2048, 1024, "", false); err != nil {
			panic(err)
		}
		defer part0Db.Close()

		if part1Db, err = rawdb.NewLevelDBDatabase(db1Path, 2048, 1024, "", false); err != nil {
			panic(err)
		}
		defer part1Db.Close()

		wp := workerpool.New(2)

		sliceFunc := func(partPrefix byte, sliceTarget ethdb.Database) {
			iterSlice := sourceDb.NewIterator([]byte{partPrefix}, nil)
			currBatch := sliceTarget.NewBatch()
			iterCnt := 0
			for iterSlice.Next() {
				if sliceErr := currBatch.Put(iterSlice.Key(), iterSlice.Value()); sliceErr != nil {
					panic(sliceErr)
				} else if currBatch.ValueSize() > 10*ethdb.IdealBatchSize {
					if sliceErr = currBatch.Write(); sliceErr != nil {
						panic(sliceErr)
					}
					currBatch.Reset()
				}
				iterCnt++
				if iterCnt%10000 == 0 {
					subPrefix := 0
					if len(iterSlice.Key()) > 1 {
						subPrefix = int(iterSlice.Key()[1])
					}
					report(fmt.Sprintf("slice prefix '%v.%v', iteration count %v", partPrefix, subPrefix, iterCnt))
				}
			}
			// make sure we write out the last batch
			if err = currBatch.Write(); err != nil {
				panic(err)
			}
			report(fmt.Sprintf("FINISHED slice prefix '%v', iteration count %v", partPrefix, iterCnt))
		}

		wp.Submit(func() {
			for i := 0; i < 0x80; i += 1 {
				sliceFunc(byte(i), part0Db)
			}
		})
		wp.Submit(func() {
			for i := 0x80; i < 0x100; i += 1 {
				sliceFunc(byte(i), part1Db)
			}
		})

		wp.StopWait()
	}

	return nil
}
