// handlers_test.go
package main

import (
	"testing"

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
           --data '{"message": "xyz","gasPrice": 5000}' \
           http://localhost:8080/omessage
       echo ""
       nloops=$((nloops+1))
       sleep 0.1
done
*/

var INPUTS = map[int]string{
	0: "{\"message\": \"xyz\",\"gasPrice\": 5000}",
	//...
}

//{"message": "xyz","gasPrice": 5000}
func TestInsert(t *testing.T) {
	mpDB := newMemPoolDB()
	for _, input := range INPUTS {
		txn := StringToTransaction(input)
		mpDB.insert(&txn)
	}
	for _, input := range INPUTS {
		// search and find each expected txn in mempool
	}
}

func TestDelete(t *testing.T) {
	mpDB := newMemPoolDB()
	for _, input := range INPUTS {
		txn := StringToTransaction(input)
		mpDB.insert(&txn)
	}
	toDelete := map[int]string{
		0: "{\"message\": \"xyz\",\"gasPrice\": 5000}",
		// ...
	}
	for _, item := range toDelete {
		txn := StringToTransaction(item)
		mpDB.delete(&txn)
	}
	// require.DeepEquals expected txn list in mempool
}

func TestSortGreatestFirst(t *testing.T) {
	mpDB := newMemPoolDB()
	for _, input := range INPUTS {
		txn := StringToTransaction(input)
		mpDB.insert(&txn)
	}
	require.NoError(t, mpDB.sortGreatestFirst())
	// require.DeepEquals expected list

	// error case?

}

func TestMakeBlock(t *testing.T) {
	mpDB := newMemPoolDB()
	for _, input := range INPUTS {
		txn := StringToTransaction(input)
		mpDB.insert(&txn)
	}
	blk, err := mpDB.makeBlock()
	require.NoError(t, err)
	// check block
	// check mempool left

	// error case?

}
