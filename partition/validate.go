package partition

import (
	"encoding/hex"
	"fmt"
)

func ValidatePOCPartitionedDatabase(dbPartPaths []string, freezerPath string, report func(string)) error {

	if validDb, err := NewPOCPartitionedDatabaseWithFreezer(dbPartPaths, 0, 0, freezerPath, ""); err != nil {
		return err
	} else {
		iterCnt := 0
		iter := validDb.NewIterator(nil, nil)
		for iter.Next() {
			iterCnt++
			if iterCnt%10000 == 0 {
				report(fmt.Sprintf("iteration %v, key '%v'", iterCnt, hex.EncodeToString(iter.Key())))
			}
		}
	}

	return nil
}
