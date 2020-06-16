package oserver

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bash script to generate txns
/*
#!/bin/bash
nloops=1
while [ $nloops -le 5 ]
do
       echo "Sending a message at: $(($(gdate +%s%N)/1000000))"
       curl --header "Content-Type: application/json" \
           --request POST \
           --data '{"message": "xyz","gasprice": 5000}' \
           http://localhost:8080/omessage
       echo ""
       nloops=$((nloops+1))
       sleep 0.1
done
*/

func TestInsert(t *testing.T) {
	mpDB := newMemPoolDB()
	for _, input := range TEST_INPUTS {
		txn := new(Transaction)
		err := txn.MessageToTransaction([]byte(input))
		require.NoError(t, err)
		mpDB.insert(txn)
	}

	// check
	if len(mpDB.mempool) != len(TEST_INPUTS) {
		t.Fatalf("mempool has (%v) != (%v)", len(mpDB.mempool), len(TEST_INPUTS))
	}
	matched := 0
	for _, actual := range mpDB.mempool {
		actualmsg, err := actual.TransactionToMessage()
		require.NoError(t, err)
		//tprint("actualmsg(%s)", string(actualmsg))
		for _, expected := range TEST_INPUTS {
			//tprint("expected(%s)", string(expected))
			if string(actualmsg) == expected {
				matched++
				break
			}
		}
	}
	if len(TEST_INPUTS) != matched {
		t.Fatalf("Insert Failed, expected(%v) != matched(%v)", len(TEST_INPUTS), matched)
	}
}

func TestSortDescending(t *testing.T) {
	mpDB := newMemPoolDB()
	for _, input := range TEST_INPUTS {
		txn := new(Transaction)
		require.NoError(t, txn.MessageToTransaction([]byte(input)))
		mpDB.insert(txn)
	}
	require.NoError(t, mpDB.sortDescending())

	// check
	for i, expected := range TEST_SORTEDINPUTS {
		actual := mpDB.mempool[i]
		actualmsg, err := actual.TransactionToMessage()
		require.NoError(t, err)
		if expected != string(actualmsg) {
			t.Fatalf("SortDescending Failed expected(%v) actual(%v)", expected, string(actualmsg))
		}
	}
}

func TestMakeBlock(t *testing.T) {
	mpDB := newMemPoolDB()
	for _, input := range TEST_INPUTS {
		txn := new(Transaction)
		require.NoError(t, txn.MessageToTransaction([]byte(input)))
		mpDB.insert(txn)
	}
	blk := mpDB.makeBlock(TEST_BLOCKSIZE)

	// check
	if len(blk) != TEST_BLOCKSIZE {
		t.Fatalf("MakeBlock failed, blocksize (%v) is not the right size (%v)", len(blk), TEST_BLOCKSIZE)
	}
	if len(mpDB.mempool) != len(TEST_RESTOFMEMPOOL) {
		t.Fatalf("MakeBlock failed, mempool size (%v) is not the right size (%v)", len(mpDB.mempool), len(TEST_RESTOFMEMPOOL))
	}

	for i, expected := range TEST_BLOCK {
		actual := blk[i]
		actualmsg, err := actual.TransactionToMessage()
		require.NoError(t, err)
		if expected != string(actualmsg) {
			t.Fatalf("MakeBlock Failed expected(%v) actual(%v)", expected, string(actualmsg))
		}
	}
	for i, expected := range TEST_RESTOFMEMPOOL {
		actual := mpDB.mempool[i]
		actualmsg, err := actual.TransactionToMessage()
		require.NoError(t, err)
		if expected != string(actualmsg) {
			t.Fatalf("MakeBlock mempool Failed expected(%v) actual(%v)", expected, string(actualmsg))
		}
	}

	// check smaller block than size allowed
	blockSizeAllowed := 10
	smallBlock := mpDB.makeBlock(blockSizeAllowed)
	assert.Equal(t, len(smallBlock), len(TEST_RESTOFMEMPOOL))
	assert.Equal(t, 0, len(mpDB.mempool))

	// check empty mempool
	mpDB.mempool = make([]*Transaction, 0)
	emptyBlock := mpDB.makeBlock(TEST_BLOCKSIZE)
	assert.Equal(t, 0, len(emptyBlock))

}

func TestBlockServer(t *testing.T) {

	oserver := newOServer(TEST_BLOCKSIZE, TEST_BLOCKMILLISECONDS, TEST_PORT)
	running := oserver.Run()
	if !running {
		t.Fatal("bad parameters")
	}

	seed := rand.NewSource(time.Now().UnixNano())
	rd := rand.New(seed)
	client := &http.Client{}
	// todo: randomized input messages + gasprices
	for _, input := range TEST_INPUTS {
		req, err := http.NewRequest("POST", TEST_URL, bytes.NewBuffer([]byte(input)))
		if err != nil {
			t.Fatalf("Error reading request. %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatal("status not ok")
		}
		tprint("sent req(%v)", input)
		time.Sleep(time.Millisecond * time.Duration(rd.Intn(100)))
	}

	// try a msg that is not a txn
	badmsg := "{\"something\":\"xyz\",\"someotherthing\":5000}"
	tprint("badmsg %s", badmsg)
	req, err := http.NewRequest("POST", TEST_URL, bytes.NewBuffer([]byte(badmsg)))
	if err != nil {
		t.Fatalf("Error reading request. %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("badmsg should not be ok")
	}

}

func TestIsValid(t *testing.T) {
	oserver := newOServer(-2, TEST_BLOCKMILLISECONDS, TEST_PORT)
	valid := oserver.Run()
	if valid {
		t.Fatal("not caught: bad blocksize")
	}
	oserver = newOServer(TEST_BLOCKSIZE, -100, TEST_PORT)
	valid = oserver.Run()
	if valid {
		t.Fatal("not caught: bad blockmilliseconds")
	}
	oserver = newOServer(TEST_BLOCKSIZE, TEST_BLOCKMILLISECONDS, -1000)
	valid = oserver.Run()
	if valid {
		t.Fatal("not caught: bad port")
	}
}

var TEST_INPUTS = []string{
	0:  "{\"message\":\"xyz\",\"gasprice\":5000}",
	1:  "{\"message\":\"hello\",\"gasprice\":40}",
	2:  "{\"message\":\"abc\",\"gasprice\":5050}",
	3:  "{\"message\":\"222\",\"gasprice\":5001}",
	4:  "{\"message\":\"this gas price is the best\",\"gasprice\":0}",
	5:  "{\"message\":\"avery\",\"gasprice\":22}",
	6:  "{\"message\":\"myTransaction\",\"gasprice\":800}",
	7:  "{\"message\":\"alice 234\",\"gasprice\":100000}",
	8:  "{\"message\":\"123 bob\",\"gasprice\":1}",
	9:  "{\"message\":\"aaaaaaa\",\"gasprice\":16}",
	10: "{\"message\":\"whodis\",\"gasprice\":333}",
}

var TEST_BLOCKSIZE = 3
var TEST_BLOCKMILLISECONDS = int64(200)
var TEST_PORT = 8080
var TEST_URL = "http://localhost:8080"
var TEST_BLOCK = []string{
	0: "{\"message\":\"alice 234\",\"gasprice\":100000}",
	1: "{\"message\":\"abc\",\"gasprice\":5050}",
	2: "{\"message\":\"222\",\"gasprice\":5001}",
}

var TEST_RESTOFMEMPOOL = []string{
	0: "{\"message\":\"xyz\",\"gasprice\":5000}",
	1: "{\"message\":\"myTransaction\",\"gasprice\":800}",
	2: "{\"message\":\"whodis\",\"gasprice\":333}",
	3: "{\"message\":\"hello\",\"gasprice\":40}",
	4: "{\"message\":\"avery\",\"gasprice\":22}",
	5: "{\"message\":\"aaaaaaa\",\"gasprice\":16}",
	6: "{\"message\":\"123 bob\",\"gasprice\":1}",
	7: "{\"message\":\"this gas price is the best\",\"gasprice\":0}",
}

var TEST_SORTEDINPUTS = []string{
	0:  "{\"message\":\"alice 234\",\"gasprice\":100000}",
	1:  "{\"message\":\"abc\",\"gasprice\":5050}",
	2:  "{\"message\":\"222\",\"gasprice\":5001}",
	3:  "{\"message\":\"xyz\",\"gasprice\":5000}",
	4:  "{\"message\":\"myTransaction\",\"gasprice\":800}",
	5:  "{\"message\":\"whodis\",\"gasprice\":333}",
	6:  "{\"message\":\"hello\",\"gasprice\":40}",
	7:  "{\"message\":\"avery\",\"gasprice\":22}",
	8:  "{\"message\":\"aaaaaaa\",\"gasprice\":16}",
	9:  "{\"message\":\"123 bob\",\"gasprice\":1}",
	10: "{\"message\":\"this gas price is the best\",\"gasprice\":0}",
}

func tprint(in string, args ...interface{}) {
	if in == "\n" {
		fmt.Println()
	} else {
		fmt.Printf("[test] "+in+"\n", args...)
	}
}
