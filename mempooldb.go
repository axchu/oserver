package oserver

import (
	"sort"
	"sync"
)

type memPoolDB struct {
	sync.Mutex
	mempool []*Transaction
}

func newMemPoolDB() *memPoolDB {
	mpDB := new(memPoolDB)
	mpDB.mempool = []*Transaction{}
	return mpDB
}

func (mpDB *memPoolDB) insert(txn *Transaction) {
	mpDB.Lock()
	defer mpDB.Unlock()
	mpDB.mempool = append(mpDB.mempool, txn)
}

// to use pkg sort's nlogn optimized quicksort
type ByGasPrice []*Transaction

func (gpList ByGasPrice) Len() int           { return len(gpList) }
func (gpList ByGasPrice) Less(i, j int) bool { return gpList[i].GasPrice < gpList[j].GasPrice }
func (gpList ByGasPrice) Swap(i, j int)      { gpList[i], gpList[j] = gpList[j], gpList[i] }

func (mpDB *memPoolDB) sortDescending() (err error) {
	mpDB.Lock()
	defer mpDB.Unlock()

	sort.Sort(sort.Reverse(ByGasPrice(mpDB.mempool)))

	return nil
}

func (mpDB *memPoolDB) makeBlock(blockSize int) (block []*Transaction) {
	mpDB.sortDescending()
	blockLength := 0

	mpDB.Lock()
	defer mpDB.Unlock()

	memPoolLength := len(mpDB.mempool)
	switch {
	case memPoolLength == 0:
		blockLength = 0
	case memPoolLength < blockSize:
		blockLength = memPoolLength
	default:
		blockLength = blockSize
	}
	blk := mpDB.mempool[0:blockLength]
	mpDB.mempool = mpDB.mempool[blockLength:memPoolLength]
	return blk
}
