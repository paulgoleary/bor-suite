package partition

import (
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

type kv struct {
	k []byte
	v []byte
}

type testIterator struct {
	d []kv
	s bool
}

func makeRandBytes() []byte {
	b := make([]byte, rand.Intn(20))
	rand.Read(b)
	return b
}
func makeTestIter(sz int) *testIterator {
	d := make([]kv, sz)
	for i := 0; i < sz; i++ {
		d[i].k = append(d[i].k, byte(i))
		d[i].k = append(d[i].k, makeRandBytes()...)
		d[i].v = append(d[i].v, makeRandBytes()...)
	}

	return &testIterator{
		d: d,
		s: false,
	}
}

func (t *testIterator) copy() *testIterator {
	return &testIterator{t.d, t.s}
}

func (t *testIterator) Next() bool {
	if len(t.d) == 0 {
		return false
	}
	if !t.s {
		t.s = true
		return true
	} else {
		t.d = t.d[1:]
		return len(t.d) != 0
	}
}

func (t *testIterator) Error() error {
	return nil
}

func (t *testIterator) Key() []byte {
	if len(t.d) == 0 {
		panic("empty iterator")
	}
	return t.d[0].k
}

func (t *testIterator) Value() []byte {
	if len(t.d) == 0 {
		panic("empty iterator")
	}
	return t.d[0].v
}

func (t *testIterator) Release() {
}

var _ ethdb.Iterator = &testIterator{}

func TestEthIteratorCompare(t *testing.T) {

	il := makeTestIter(10)
	ir := il.copy()

	sz, err := compareIters(il, ir)
	assert.Equal(t, nil, err)
	assert.True(t, sz > 0)

	il = makeTestIter(10)
	ir = makeTestIter(20)

	_, err = compareIters(il, ir)
	assert.NotEqual(t, nil, err)

	// TODO: hmmmm... not a very good test yet...

}
