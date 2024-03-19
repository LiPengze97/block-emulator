// Definition of block

package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
)

// The definition of block
type JakiroBlock struct {
	Header  *BlockHeader
	Header2 *BlockHeader
	Body    []*Transaction
	Body2   []*Transaction // for another chain
	Hash    []byte
}

func NewJakiroBlock(bh *BlockHeader, bb []*Transaction, bh2 *BlockHeader, bb2 []*Transaction) *JakiroBlock {
	return &JakiroBlock{Header: bh, Body: bb, Header2: bh2, Body2: bb2}
}

func (b *JakiroBlock) PrintJakiroBlock() string {
	vals := []interface{}{
		b.Header.Number,
		b.Hash,
	}
	res := fmt.Sprintf("%v\n", vals)
	fmt.Println(res)
	return res
}

// Encode JakiroBlock for storing
func (b *JakiroBlock) Encode() []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// Decode JakiroBlock
func DecodeJB(b []byte) *JakiroBlock {
	var block JakiroBlock

	decoder := gob.NewDecoder(bytes.NewReader(b))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
