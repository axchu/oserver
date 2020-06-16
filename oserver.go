package oserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type Transaction struct {
	Message  string `json:"message"`
	GasPrice int64  `json:"gasprice"`
}

type memPoolDB struct {
	sync.Mutex
	mempool   []*Transaction
	blockSize int
}

func newMemPoolDB(blockSize int) *memPoolDB {
	mpDB := new(memPoolDB)
	mpDB.mempool = []*Transaction{}
	mpDB.blockSize = blockSize
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

func (mpDB *memPoolDB) makeBlock() (block []*Transaction) {
	mpDB.sortDescending()

	mpDB.Lock()
	defer mpDB.Unlock()

	blockLength := 0
	memPoolLength := len(mpDB.mempool)
	switch {
	case memPoolLength == 0:
		//log.Printf("memPoolDB:makeBlock nothing in mempool. No block made.")
		blockLength = 0
	case memPoolLength < mpDB.blockSize:
		blockLength = memPoolLength
	default:
		blockLength = mpDB.blockSize
	}
	blk := mpDB.mempool[0:blockLength]
	mpDB.mempool = mpDB.mempool[blockLength:memPoolLength]
	return blk
}

func (mpDB *memPoolDB) BlockServer(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	txn, err := MessageToTransaction(body)
	if err != nil {
		http.Error(w, "message "+string(body)+" is not transaction", http.StatusBadRequest)
		return
	}
	mpDB.insert(&txn)
	//log.Printf("inserted(%+v)\n", txn)
}

func MessageToTransaction(message []byte) (txn Transaction, err error) {
	err = json.Unmarshal(message, &txn)
	if txn.Message == "" && txn.GasPrice == 0 {
		err = fmt.Errorf("ERROR MessageToTransaction: not a transaction (%s)", string(message))
		//log.Printf(err.Error())
	}
	return txn, err
}

func TransactionToMessage(txn Transaction) (message []byte, err error) {
	message, err = json.Marshal(txn)
	// if err != nil {
	// 	log.Printf("ERROR TransactionToMessage %v", err)
	// }
	return message, err
}

func Init(blockSize int, blockMilliseconds int64, port string) {

	mpDB := newMemPoolDB(blockSize)
	go func() {
		http.HandleFunc("/", mpDB.BlockServer)
		http.ListenAndServe(port, nil)
	}()

	go func() {
		for {
			time.Sleep(time.Millisecond * time.Duration(blockMilliseconds))
			block := mpDB.makeBlock()
			blockString := BlockToString(block)
			log.Printf("block made(\n%v)", blockString)
		}
	}()

}

// BlockToString is a helper print function
func BlockToString(block []*Transaction) (str string) {
	str = ""
	for _, txn := range block {
		msg, err := TransactionToMessage(*txn)
		if err != nil {
			panic(err)
		}
		str += "[" + string(msg) + "]\n"
	}
	return str
}
