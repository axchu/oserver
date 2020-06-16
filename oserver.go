package oserver

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type OServer struct {
	blockSize         int
	blockMilliseconds int64
	port              int
	mpDB              *memPoolDB
}

func newOServer(blockSize int, blockMilliseconds int64, port int) (oserver *OServer) {
	oserver = new(OServer)
	oserver.mpDB = newMemPoolDB()
	oserver.blockSize = blockSize
	oserver.blockMilliseconds = blockMilliseconds
	oserver.port = port
	return oserver

}

func (oserver *OServer) Run() bool {

	if !oserver.IsValid() {
		return false
	}
	go func() {
		http.HandleFunc("/", oserver.BlockServer)
		http.ListenAndServe(":"+strconv.Itoa(oserver.port), nil)
	}()

	go func() {
		for {
			time.Sleep(time.Millisecond * time.Duration(oserver.blockMilliseconds))
			block := oserver.mpDB.makeBlock(oserver.blockSize)
			blockString := BlockToString(block)
			log.Printf("block made(\n%v)", blockString)
		}
	}()
	// todo: more gracefully close using context
	return true
}

func (oserver *OServer) IsValid() bool {
	valid := true
	if oserver.blockSize <= 0 { // todo: maxblocksize
		log.Printf("ERR: invalid blocksize(%v)", oserver.blockSize)
		valid = false
	}
	if oserver.blockMilliseconds <= 0 { // todo: maxBlockMilliseconds
		log.Printf("ERR: invalid blockMilliseconds(%v)", oserver.blockMilliseconds)
		valid = false
	}

	//NOTE: 65535 is max port
	if oserver.port <= 0 || oserver.port > 65535 {
		log.Printf("ERR: invalid port(%v)", oserver.port)
		valid = false
	}

	return valid
}

func (oserver *OServer) BlockServer(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	txn := new(Transaction)
	err = txn.MessageToTransaction(body)
	if err != nil {
		http.Error(w, "message "+string(body)+" is not transaction", http.StatusBadRequest)
		return
	}
	oserver.mpDB.insert(txn)
}

// BlockToString is a helper print function
func BlockToString(block []*Transaction) (str string) {
	str = ""
	for _, txn := range block {
		msg, err := txn.TransactionToMessage()
		if err != nil {
			panic(err)
		}
		str += "[" + string(msg) + "]\n"
	}
	return str
}
