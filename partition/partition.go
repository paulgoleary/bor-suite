package main

import (
	"fmt"
	"github.com/maticnetwork/bor/core/rawdb"
	"github.com/maticnetwork/bor/ethdb"
	"os"
	"path/filepath"
)

func main() {
	argsWithoutProg := os.Args[1:]

	dbPath := argsWithoutProg[0]

	if sourceDb, err := rawdb.NewLevelDBDatabaseWithFreezer(dbPath, 0, 0, filepath.Join(dbPath, "ancient"), ""); err != nil {
		panic(err)
	} else {

		db0Path := dbPath + ".0"
		db1Path := dbPath + ".1"

		var part0Db, part1Db ethdb.Database

		if part0Db, err = rawdb.NewLevelDBDatabaseWithFreezer(db0Path, 0, 0, filepath.Join(db0Path, "ancient"), ""); err != nil {
			panic(err)
		}

		if part1Db, err = rawdb.NewLevelDBDatabaseWithFreezer(db1Path, 0, 0, filepath.Join(db1Path, "ancient"), ""); err != nil {
			panic(err)
		}

		// TODO: batching?

		iterAll := sourceDb.NewIterator(nil, nil)
		cntIter := 0
		for iterAll.Next() {
			if len(iterAll.Key()) == 0 {
				panic("should not happen - empty key value")
			}
			if iterAll.Key()[0] < 0x80 {
				if err = part0Db.Put(iterAll.Key(), iterAll.Value()); err != nil {
					panic(err)
				}
			} else {
				if err = part1Db.Put(iterAll.Key(), iterAll.Value()); err != nil {
					panic(err)
				}
			}
			cntIter++
			if cntIter%10000 == 0 {
				println(fmt.Sprintf("iteration count %v, current prefix %v", cntIter, iterAll.Key()[0]))
			}
		}
	}
}
