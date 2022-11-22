package indexer

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/umbracle/ethgo"
)

type BorLogSink interface {
	ProcessBorLogs([]*types.Log) error
}

func BorLogsToEthGo(bl []*types.Log) (el []*ethgo.Log) {
	for i := range bl {
		el = append(el, BorLogToEthGo(bl[i]))
	}
	return
}

func BorLogToEthGo(bl *types.Log) *ethgo.Log {
	var tps []ethgo.Hash
	for x := range bl.Topics {
		tps = append(tps, ethgo.Hash(bl.Topics[x]))
	}
	return &ethgo.Log{
		Removed:          bl.Removed,
		LogIndex:         uint64(bl.Index),
		TransactionIndex: uint64(bl.TxIndex),
		TransactionHash:  ethgo.Hash(bl.TxHash),
		BlockHash:        ethgo.Hash(bl.BlockHash),
		BlockNumber:      bl.BlockNumber,
		Address:          ethgo.Address(bl.Address),
		Topics:           tps,
		Data:             bl.Data,
	}
}
