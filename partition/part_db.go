package partition

import (
	"fmt"
	"github.com/maticnetwork/bor/core/rawdb"
	"github.com/maticnetwork/bor/ethdb"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
)

func bytePartition(b byte, n int) int {
	return int(b) / (256 / n)
}

type partTable []ethdb.Database

func (pt partTable) get(b byte) ethdb.Database {
	return pt[bytePartition(b, len(pt))]
}

type PartitionedDatabase struct {
	pt partTable
}

var _ ethdb.Database = &PartitionedDatabase{}

func getPartNameOrd(name string) (ord int, err error) {
	s := strings.Split(name, ".")
	if len(s) != 2 {
		err = fmt.Errorf("should not happen - part name is not valid")
		return
	}
	if ord, err = strconv.Atoi(s[1]); err != nil {
		err = fmt.Errorf("should not happen - part ord is not valid integer")
		return
	}
	return
}

func isPartName(name string) bool {
	_, err := getPartNameOrd(name)
	return err == nil
}

func findPartitions(partRoot string) (ret []string, err error) {

	partDir, partName := filepath.Split(partRoot)

	walkFunc := func(path string, dir fs.DirEntry, err error) error {
		if dir.IsDir() && dir.Name() != partName {
			if isPartName(dir.Name()) {
				ret = append(ret, filepath.Join(partDir, dir.Name()))
			}
		}
		return nil
	}

	if err = filepath.WalkDir(partRoot, walkFunc); err != nil {
		return
	}

	return
}

func NewPOCPartitionedDatabaseWithFreezer(dbRoot string, cache int, handles int, freezerName string, namespace string) (db ethdb.Database, err error) {

	var partDirs []string
	if partDirs, err = findPartitions(dbRoot); err != nil {
		return
	}
	if len(partDirs) != 2 {
		err = fmt.Errorf("should not happen - POC version only supports two partitions")
		return
	}

	pt := make([]ethdb.Database, len(partDirs))

	for _, d := range partDirs {
		_, n := filepath.Split(d)
		var ord int
		if ord, err = getPartNameOrd(n); err != nil {
			return
		}
		if ord >= len(partDirs) {
			err = fmt.Errorf("should not happen - invalid partition ordinal: %v", n)
			return
		}
		if pt[ord], err = rawdb.NewLevelDBDatabaseWithFreezer(d, cache, handles, filepath.Join(d, freezerName), namespace); err != nil {
			return
		}
	}

	return &PartitionedDatabase{pt: partTable(pt)}, nil
}

func (p PartitionedDatabase) Has(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, nil // TODO: is this correct?
	}
	return p.pt.get(key[0]).Has(key)
}

func (p PartitionedDatabase) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, nil // TODO: is this correct?
	}
	return p.pt.get(key[0]).Get(key)
}

func (p PartitionedDatabase) HasAncient(kind string, number uint64) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p PartitionedDatabase) Ancient(kind string, number uint64) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (p PartitionedDatabase) Ancients() (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (p PartitionedDatabase) AncientSize(kind string) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (p PartitionedDatabase) Put(key []byte, value []byte) error {
	if len(key) == 0 {
		return fmt.Errorf("'put' of nil / empty key not supported")
	}
	return p.pt.get(key[0]).Put(key, value)
}

func (p PartitionedDatabase) Delete(key []byte) error {
	if len(key) == 0 {
		return fmt.Errorf("'delete' of nil / empty key not supported")
	}
	return p.pt.get(key[0]).Delete(key)
}

func (p PartitionedDatabase) AppendAncient(number uint64, hash, header, body, receipt, td []byte) error {
	//TODO implement me
	panic("implement me")
}

func (p PartitionedDatabase) TruncateAncients(n uint64) error {
	//TODO implement me
	panic("implement me")
}

func (p PartitionedDatabase) Sync() error {
	for i := range p.pt {
		if err := p.pt[i].Sync(); err != nil {
			return err
		}
	}
	return nil
}

func (p PartitionedDatabase) NewBatch() ethdb.Batch {
	//TODO implement me
	panic("implement me")
}

type partIterator struct {
	pi []ethdb.Iterator
}

// simple / naive initial impl
// TODO: will this require locking?

func (p *partIterator) Next() bool {
	if len(p.pi) == 0 {
		return false
	}
	if n := p.pi[0].Next(); n {
		return true
	} else {
		p.pi = p.pi[1:]
		return p.Next()
	}
}

func (p *partIterator) Error() error {
	if len(p.pi) == 0 {
		return nil
	}
	return p.pi[0].Error()
}

func (p *partIterator) Key() []byte {
	if len(p.pi) == 0 {
		return nil
	}
	return p.pi[0].Key()
}

func (p *partIterator) Value() []byte {
	if len(p.pi) == 0 {
		return nil
	}
	return p.pi[0].Value()
}

func (p *partIterator) Release() {
	for i := range p.pi {
		p.pi[i].Release()
	}
}

var _ ethdb.Iterator = &partIterator{}

func (p PartitionedDatabase) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	pi := &partIterator{}
	for i := range p.pt {
		// TODO: since each partition is disjoint and - by definition - byte sequential this *should* work
		//  but more thought a/o testing likely required :)
		pi.pi = append(pi.pi, p.pt[i].NewIterator(prefix, start))
	}
	return pi
}

func (p PartitionedDatabase) Stat(property string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p PartitionedDatabase) Compact(start []byte, limit []byte) error {
	//TODO implement me
	panic("implement me")
}

// TODO: deal with partial failures and such ...?
func (p PartitionedDatabase) Close() error {
	for i := range p.pt {
		if err := p.pt[i].Close(); err != nil {
			return err
		}
	}
	return nil
}
