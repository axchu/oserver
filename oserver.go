package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

var BLOCKSIZE = int64(5)

type memPoolDB struct {
	sync.Mutex
	mempool   []*Transaction
	blockSize int64
}

type block []*Transaction

type Transaction struct {
	Message  string `json:"message"`
	GasPrice int64  `json:"gasprice"`
}

func newMemPoolDB() *memPoolDB {
	mpDB := new(memPoolDB)
	mpDB.mempool = []*Transaction{}
	mpDB.blockSize = BLOCKSIZE
	return mpDB
}

func (mpDB *memPoolDB) insert(txn *Transaction) {
	mpDB.Lock()
	defer mpDB.Unlock()

}

func (mpDB *memPoolDB) delete(txn *Transaction) {
	mpDB.Lock()
	defer mpDB.Unlock()
}

func (mpDB *memPoolDB) sortGreatestFirst() (err error) {
	mpDB.Lock()
	defer mpDB.Unlock()
	return nil
}

func (mpDB *memPoolDB) makeBlock() (blk *block, err error) {
	//mpDB.Lock()
	//defer mpDB.Unlock()

	// sort greatest first

	// copy top [0:blockSize] out, delete it from the mempool

	return blk, nil
}

func StringToTransaction(message string) (txn Transaction) {
	err := json.Unmarshal([]byte(message), &txn)
	if err != nil {
		log.Fatal(err)
	}
	return txn
}

func main() {

	http.HandleFunc("/", HelloServer)
	http.ListenAndServe(":8080", nil)

}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	log.Printf("%s\n", body)

	// add txn to mempool

	// every 3 seconds, make a block; write the block to log and the

}
