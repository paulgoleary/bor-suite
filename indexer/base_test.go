package indexer

import (
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
	"testing"
)

func TestEventCounterSink(t *testing.T) {

	sink := MakeEventCounterSink("Transfer")
	require.NotNil(t, sink)

	transEvent, ok := erc721ABI.Events["Transfer"]
	require.True(t, ok)

	testTopics := []ethgo.Hash{transEvent.ID()}
	logs := []*ethgo.Log{{Topics: testTopics}}
	sink.Process(logs)

	require.Equal(t, uint64(1), sink.Count.Load())
}
