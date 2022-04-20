package partition

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestPartitionDir(t *testing.T) {

	dir, err := ioutil.TempDir("", "test-partition-*")
	assert.Equal(t, nil, err)
	defer os.RemoveAll(dir)

	os.MkdirAll(filepath.Join(dir, "chaindata"), 777)
	os.MkdirAll(filepath.Join(dir, "chaindata.0"), 777)
	os.MkdirAll(filepath.Join(dir, "chaindata.1"), 777)
	os.MkdirAll(filepath.Join(dir, "some_trash"), 777)

	maybeParts, err := findPartitions(dir)
	assert.Equal(t, nil, err)
	assert.True(t, len(maybeParts) > 0)

	// TODO: more validation ...?
}

func TestBytePartitioning(t *testing.T) {
	assert.Equal(t, 0, bytePartition(0x00, 2))
	assert.Equal(t, 0, bytePartition(0x7F, 2))
	assert.Equal(t, 1, bytePartition(0x80, 2))
	assert.Equal(t, 1, bytePartition(0xFF, 2))

	// TODO: more part numbers ...?
}
