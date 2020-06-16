package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

var BLOCKSIZE = 5

type memPoolDB struct {
	sync.Mutex
	mempool   []*Transaction
	blockSize int
}

type Transaction struct {
	Message  string `json:"message"`
	GasPrice int64  `json:"gasprice"`
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
	//dprint("mempool")
	//dprint(BlockToString(mpDB.mempool))
	blk := mpDB.mempool[0:blockLength]
	//dprint("block")
	//dprint(BlockToString(blk))
	mpDB.mempool = mpDB.mempool[blockLength:memPoolLength]
	//dprint("new mempool")
	//dprint(BlockToString(mpDB.mempool))
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
	log.Printf("inserted(%+v)\n", txn)
	// ctx := r.Context()
	// select {
	// case <-time.After(3 * time.Second):
	//
	// case <-ctx.Done():
	// 	err := ctx.Err()
	// 	fmt.Println("server:", err)
	// 	internalError := http.StatusInternalServerError
	// 	http.Error(w, err.Error(), internalError)
	// }
}

func MessageToTransaction(message []byte) (txn Transaction, err error) {
	err = json.Unmarshal(message, &txn)
	// if err != nil {
	// 	log.Printf("ERROR MessageToTransaction %v", err)
	// }
	return txn, err
}

func TransactionToMessage(txn Transaction) (message []byte, err error) {
	message, err = json.Marshal(txn)
	// if err != nil {
	// 	log.Printf("ERROR TransactionToMessage %v", err)
	// }
	return message, err
}

func Init() {

	mpDB := newMemPoolDB(BLOCKSIZE)
	go func() {
		log.Printf("server coming up\n")
		http.HandleFunc("/", mpDB.BlockServer)
		http.ListenAndServe(":8080", nil)
	}()

	go func() {
		log.Printf("making block goroutine spawned\n")
		for {
			//time.Sleep(3 * time.Second)
			time.Sleep(time.Millisecond * 100)
			block := mpDB.makeBlock()
			blockString := BlockToString(block)
			log.Printf("block made(%v)", blockString)
		}
	}()

}

func main() {
	Init()
}

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

// func dprint(in string, args ...interface{}) {
// 	if in == "\n" {
// 		fmt.Println()
// 	} else {
// 		fmt.Printf("[debug] "+in+"\n", args...)
// 	}
// }
