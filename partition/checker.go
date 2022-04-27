package partition

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/maticnetwork/bor/core/rawdb"
	"github.com/maticnetwork/bor/ethdb"
	"math/rand"
	"path/filepath"
)

func makeRandomKey() []byte {
	key := make([]byte, 2+rand.Intn(4))
	rand.Read(key)
	return key
}

func compareIters(il, ir ethdb.Iterator) (sz int, err error) {
	defer il.Release()
	defer ir.Release()
	for il.Next() {
		if !ir.Next() {
			if ir.Error() != nil {
				err = fmt.Errorf("right iter ended before left (%v, '%v'): %v", sz, hex.EncodeToString(il.Key()), err.Error())
			} else {
				err = fmt.Errorf("right iter ended before left (%v, '%v')", sz, hex.EncodeToString(il.Key()))
			}
			return
		}
		if bytes.Compare(il.Key(), ir.Key()) != 0 {
			err = fmt.Errorf("got non-equivalent keys")
			return
		}
		if bytes.Compare(il.Value(), ir.Value()) != 0 {
			err = fmt.Errorf("got non-equivalent keys")
			return
		}
		sz += len(il.Key()) + len(il.Value())
	}
	if ir.Next() {
		err = fmt.Errorf("left iter terminated before right")
	}
	return
}

func CheckPOCPartitionedDatabase(srcDbPath, freezerPath string, checkDbPartPaths []string, checks int, report func(string)) error {

	if sourceDb, err := rawdb.NewLevelDBDatabaseWithFreezer(srcDbPath, 0, 0, filepath.Join(srcDbPath, "ancient"), ""); err != nil {
		return err
	} else {
		if checkDb, err := NewPOCPartitionedDatabaseWithFreezer(checkDbPartPaths, 0, 0, freezerPath, ""); err != nil {
			return err
		} else {
			for i := 0; i < checks; i++ {
				checkKey := makeRandomKey()
				report(fmt.Sprintf("checking rando key '%v'", hex.EncodeToString(checkKey)))
				sourceIter := sourceDb.NewIterator(checkKey, nil)
				checkIter := checkDb.NewIterator(checkKey, nil)
				if sz, err := compareIters(sourceIter, checkIter); err != nil {
					return err
				} else {
					report(fmt.Sprintf("successful check iteration, %v bytes", sz))
				}
			}
		}
	}

	return nil
}
