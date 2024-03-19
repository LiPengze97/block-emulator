package utils

import (
	"encoding/json"
)

// TxRead describes a read operation on a key.
type TxRead struct {
	Key          []byte `json:"key"`
	Value        []byte `json:"value"`
	ContractName string `json:"contract_name"`
	Version      string `json:"version"`
}

// TxWrite describes a write operation on a key.
type TxWrite struct {
	Key          []byte `json:"key"`
	Value        []byte `json:"value"`
	ContractName string `json:"contract_name"`
}

// TxRWSet represents the overall transaction structure.
type TxRWSet struct {
	TxId     string     `json:"tx_id"`
	TxReads  []*TxRead  `json:"tx_reads"`
	TxWrites []*TxWrite `json:"tx_writes"`
}

// NewTxRWSet creates a new instance of TxRWSet with default values.
func NewTxRWSet() *TxRWSet {
	return &TxRWSet{}
}

// Serialize serializes the TxRWSet message to JSON.
func (m *TxRWSet) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// Deserialize deserializes the TxRWSet message from JSON.
func (m *TxRWSet) Deserialize(data []byte) error {
	return json.Unmarshal(data, m)
}

/*
32 bytes for nonce
32 bytes for balance
20 bytes for address itself
84 bytes total
*/
