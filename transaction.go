package oserver

import (
	"encoding/json"
	"fmt"
)

type Transaction struct {
	Message  string `json:"message"`
	GasPrice int64  `json:"gasprice"`
}

func (txn *Transaction) MessageToTransaction(message []byte) (err error) {
	err = json.Unmarshal(message, &txn)
	if txn.Message == "" && txn.GasPrice == 0 {
		err = fmt.Errorf("ERROR MessageToTransaction: not a transaction (%s)", string(message))
	}
	return err
}

func (txn *Transaction) TransactionToMessage() (message []byte, err error) {
	message, err = json.Marshal(txn)
	return message, err
}
