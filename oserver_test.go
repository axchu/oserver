// handlers_test.go
package main

import (
	"bytes"
	"fmt"
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

//{"message": "xyz","gasprice": 5000}
func TestInsert(t *testing.T) {
	mpDB := newMemPoolDB(TEST_BLOCKSIZE)
	for _, input := range TEST_INPUTS {
		txn, err := MessageToTransaction([]byte(input))
		require.NoError(t, err)
		mpDB.insert(&txn)
	}

	// check
	if len(mpDB.mempool) != len(TEST_INPUTS) {
		t.Fatalf("mempool has (%v) != (%v)", len(mpDB.mempool), len(TEST_INPUTS))
	}
	matched := 0
	for _, actual := range mpDB.mempool {
		actualmsg, err := TransactionToMessage(*actual)
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
	mpDB := newMemPoolDB(TEST_BLOCKSIZE)
	for _, input := range TEST_INPUTS {
		txn, err := MessageToTransaction([]byte(input))
		require.NoError(t, err)
		mpDB.insert(&txn)
	}
	require.NoError(t, mpDB.sortDescending())

	// check
	for i, expected := range TEST_SORTEDINPUTS {
		actual := mpDB.mempool[i]
		actualmsg, err := TransactionToMessage(*actual)
		require.NoError(t, err)
		if expected != string(actualmsg) {
			t.Fatalf("SortDescending Failed expected(%v) actual(%v)", expected, string(actualmsg))
		}
	}
}

func TestMakeBlock(t *testing.T) {
	mpDB := newMemPoolDB(TEST_BLOCKSIZE)
	for _, input := range TEST_INPUTS {
		txn, err := MessageToTransaction([]byte(input))
		require.NoError(t, err)
		mpDB.insert(&txn)
	}
	blk := mpDB.makeBlock()
	//tprint(BlockToString(blk))
	// check
	if len(blk) != TEST_BLOCKSIZE {
		t.Fatalf("MakeBlock failed, blocksize (%v) is not the right size (%v)", len(blk), TEST_BLOCKSIZE)
	}
	if len(mpDB.mempool) != len(TEST_RESTOFMEMPOOL) {
		t.Fatalf("MakeBlock failed, mempool size (%v) is not the right size (%v)", len(mpDB.mempool), len(TEST_RESTOFMEMPOOL))
	}

	for i, expected := range TEST_BLOCK {
		actual := blk[i]
		actualmsg, err := TransactionToMessage(*actual)
		require.NoError(t, err)
		if expected != string(actualmsg) {
			t.Fatalf("MakeBlock Failed expected(%v) actual(%v)", expected, string(actualmsg))
		}
	}
	for i, expected := range TEST_RESTOFMEMPOOL {
		actual := mpDB.mempool[i]
		actualmsg, err := TransactionToMessage(*actual)
		require.NoError(t, err)
		if expected != string(actualmsg) {
			t.Fatalf("MakeBlock mempool Failed expected(%v) actual(%v)", expected, string(actualmsg))
		}
	}

	// check smaller block
	mpDB.blockSize = 10
	smallBlock := mpDB.makeBlock()
	assert.Equal(t, len(smallBlock), len(TEST_RESTOFMEMPOOL))
	assert.Equal(t, 0, len(mpDB.mempool))

	// check empty mempool
	mpDB.mempool = make([]*Transaction, 0)
	emptyBlock := mpDB.makeBlock()
	assert.Equal(t, 0, len(emptyBlock))

}

func TestBlockServer(t *testing.T) {

	Init()

	for _, input := range TEST_INPUTS {
		time.Sleep(time.Millisecond * 50)
		req, err := http.NewRequest("POST", "http://localhost:8080", bytes.NewBuffer([]byte(input)))
		if err != nil {
			t.Fatalf("Error reading request. %v", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		// Check the status code is what we expect.
		//if status := rr.Code; status != http.StatusOK {
		//	t.Errorf("handler returned wrong status code: got %v want %v",
		//		status, http.StatusOK)
		///}
		tprint("sent req(%v)", input)
	}
	// Check the response body is what we expect.
	// expected := `{"alive": true}`
	// if rr.Body.String() != expected {
	// 	t.Errorf("handler returned unexpected body: got %v want %v",
	// 		rr.Body.String(), expected)
	// }

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

var TEST_BLOCKSIZE = 5
var TEST_BLOCK = []string{
	0: "{\"message\":\"alice 234\",\"gasprice\":100000}",
	1: "{\"message\":\"abc\",\"gasprice\":5050}",
	2: "{\"message\":\"222\",\"gasprice\":5001}",
	3: "{\"message\":\"xyz\",\"gasprice\":5000}",
	4: "{\"message\":\"myTransaction\",\"gasprice\":800}",
}

var TEST_RESTOFMEMPOOL = []string{
	0: "{\"message\":\"whodis\",\"gasprice\":333}",
	1: "{\"message\":\"hello\",\"gasprice\":40}",
	2: "{\"message\":\"avery\",\"gasprice\":22}",
	3: "{\"message\":\"aaaaaaa\",\"gasprice\":16}",
	4: "{\"message\":\"123 bob\",\"gasprice\":1}",
	5: "{\"message\":\"this gas price is the best\",\"gasprice\":0}",
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
