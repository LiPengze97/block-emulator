// the define and some operation of txpool

package core

import (
	"blockEmulator/utils"
	"sync"
	"time"
)

type TxPool struct {
	TxQueue              []*Transaction            // transaction Queue
	RelayPool            map[uint64][]*Transaction //designed for sharded blockchain, from Monoxide
	JakiroPool           []*Transaction
	IsJakiroTxSourcePool bool //负载高的为true，负载低的为false
	lock                 sync.Mutex
	jakiroLock           sync.Mutex
	TxNumMap             map[string]*utils.AccountTxNum // 每个地址里面有多少交易
	AddrDivsionMap       map[string]bool                // 这个地址被分到第一个部分还是第二个部分
	Part1Num             uint64
	Part2Num             uint64
	// The pending list is ignored
}

func NewTxPool() *TxPool {
	return &TxPool{
		TxQueue:              make([]*Transaction, 0),
		RelayPool:            make(map[uint64][]*Transaction),
		TxNumMap:             make(map[string]*utils.AccountTxNum),
		AddrDivsionMap:       make(map[string]bool),
		IsJakiroTxSourcePool: false,
		Part1Num:             0,
		Part2Num:             0,
	}
}

// Add a transaction to the pool (consider the queue only)
func (txpool *TxPool) AddTx2Pool(tx *Transaction) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	if tx.Time.IsZero() {
		tx.Time = time.Now()
	}
	txpool.TxQueue = append(txpool.TxQueue, tx)
	//TODO:统计一下时间开开销
	if txpool.IsJakiroTxSourcePool {
		txpool.updateTxNumMap(tx)
	}

}

func getAddrAndType(tx *Transaction) (string, int) {
	if tx.Relayed {
		return tx.Recipient, 0 // Relay2
	} else if utils.Addr2Shard(tx.Sender) == utils.Addr2Shard(tx.Recipient) {
		return tx.Sender, 1 // Intra
	} else {
		return tx.Sender, 2 // Relay1
	}
}

// 判断如何更新TxNumMap
func (txpool *TxPool) updateTxNumMap(tx *Transaction) {
	addr, type_ := getAddrAndType(tx)
	if _, ok := txpool.AddrDivsionMap[addr]; !ok {
		if txpool.Part1Num > txpool.Part2Num { // false是第二部分，true是第一部分
			txpool.AddrDivsionMap[addr] = false
		} else {
			txpool.AddrDivsionMap[addr] = true
		}
		txpool.TxNumMap[addr] = utils.NewAccountTxNum()
	}
	switch type_ {
	case 0:
		txpool.TxNumMap[addr].Relay2Tx += 1
	case 1:
		txpool.TxNumMap[addr].IntraTx += 1
	case 2:
		txpool.TxNumMap[addr].Relay1Tx += 1
	}

	if txpool.AddrDivsionMap[addr] {
		txpool.Part1Num += 1
	} else {
		txpool.Part2Num += 1
	}

}

// 减少TxNum的函数
func (txpool *TxPool) minusTxNumMap(type_ int, addr string) {
	switch type_ {
	case 0:
		txpool.TxNumMap[addr].Relay2Tx -= 1
	case 1:
		txpool.TxNumMap[addr].IntraTx -= 1
	case 2:
		txpool.TxNumMap[addr].Relay1Tx -= 1
	}

	if txpool.AddrDivsionMap[addr] {
		txpool.Part1Num -= 1
	} else {
		txpool.Part2Num -= 1
	}

	if txpool.TxNumMap[addr].Relay1Tx+txpool.TxNumMap[addr].Relay2Tx+txpool.TxNumMap[addr].IntraTx == 0 {
		delete(txpool.TxNumMap, addr)
		delete(txpool.AddrDivsionMap, addr)
	}

}

// Add a list of transactions to the pool
func (txpool *TxPool) AddTxs2Pool(txs []*Transaction) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	for _, tx := range txs {
		if tx.Time.IsZero() {
			tx.Time = time.Now()
		}
		txpool.TxQueue = append(txpool.TxQueue, tx)
		if txpool.IsJakiroTxSourcePool {
			txpool.updateTxNumMap(tx)
		}
	}
}

// add transactions into the pool head
func (txpool *TxPool) AddTxs2Pool_Head(tx []*Transaction) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	txpool.TxQueue = append(tx, txpool.TxQueue...)
}

// Pack transactions for a proposal
func (txpool *TxPool) PackTxs(max_txs uint64) []*Transaction {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	txNum := max_txs
	if uint64(len(txpool.TxQueue)) < txNum {
		txNum = uint64(len(txpool.TxQueue))
	}
	txs_Packed := txpool.TxQueue[:txNum]
	txpool.TxQueue = txpool.TxQueue[txNum:]
	return txs_Packed
}

// Pack transactions from second mempool for a proposal
func (txpool *TxPool) PackJakiroTxs(max_txs uint64) []*Transaction {
	txpool.jakiroLock.Lock()
	defer txpool.jakiroLock.Unlock()
	txNum := max_txs
	if uint64(len(txpool.TxQueue)) < txNum {
		txNum = uint64(len(txpool.TxQueue))
	}
	txs_Packed := txpool.TxQueue[:txNum]
	txpool.TxQueue = txpool.TxQueue[txNum:]
	if txpool.IsJakiroTxSourcePool {
		for i := 0; i < len(txs_Packed); i++ {
			addr, type_ := getAddrAndType(txs_Packed[i])
			txpool.minusTxNumMap(type_, addr)
		}
	}
	return txs_Packed
}

// Relay transactions
func (txpool *TxPool) AddRelayTx(tx *Transaction, shardID uint64) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	_, ok := txpool.RelayPool[shardID]
	if !ok {
		txpool.RelayPool[shardID] = make([]*Transaction, 0)
	}
	txpool.RelayPool[shardID] = append(txpool.RelayPool[shardID], tx)
}

// Split transactions for jakiro tx migrations
func (txpool *TxPool) SplitTxs() []*Transaction {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()

	// 简易版本，对半分，启用下面四行代码即可
	// txNum := uint64(len(txpool.TxQueue)) / 2
	// part1 := txpool.TxQueue[:txNum]
	// txpool.TxQueue = txpool.TxQueue[txNum:]

	// return part1

	// 细致分类

	part1 := make([]*Transaction, 0)
	part2 := make([]*Transaction, 0)

	for i := 0; i < len(txpool.TxQueue); i++ {
		tx := txpool.TxQueue[i]
		addr, type_ := getAddrAndType(tx)
		if txpool.AddrDivsionMap[addr] {
			part1 = append(part1, tx)
		} else {
			part2 = append(part2, tx)
			txpool.minusTxNumMap(type_, addr)
		}
	}
	txpool.TxQueue = part1
	return part2
}

// txpool get locked
func (txpool *TxPool) GetLocked() {
	txpool.lock.Lock()
}

// txpool get unlocked
func (txpool *TxPool) GetUnlocked() {
	txpool.lock.Unlock()
}

// get the length of tx queue
func (txpool *TxPool) GetTxQueueLen() int {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	return len(txpool.TxQueue)
}

// get the length of ClearRelayPool
func (txpool *TxPool) ClearRelayPool() {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	txpool.RelayPool = nil
}

// abort ! Pack relay transactions from relay pool
func (txpool *TxPool) PackRelayTxs(shardID, minRelaySize, maxRelaySize uint64) ([]*Transaction, bool) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	if _, ok := txpool.RelayPool[shardID]; !ok {
		return nil, false
	}
	if len(txpool.RelayPool[shardID]) < int(minRelaySize) {
		return nil, false
	}
	txNum := maxRelaySize
	if uint64(len(txpool.RelayPool[shardID])) < txNum {
		txNum = uint64(len(txpool.RelayPool[shardID]))
	}
	relayTxPacked := txpool.RelayPool[shardID][:txNum]
	txpool.RelayPool[shardID] = txpool.RelayPool[shardID][txNum:]
	return relayTxPacked, true
}

// abort ! Transfer transactions when re-sharding
func (txpool *TxPool) TransferTxs(addr utils.Address) []*Transaction {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	txTransfered := make([]*Transaction, 0)
	newTxQueue := make([]*Transaction, 0)
	for _, tx := range txpool.TxQueue {
		if tx.Sender == addr {
			txTransfered = append(txTransfered, tx)
		} else {
			newTxQueue = append(newTxQueue, tx)
		}
	}
	newRelayPool := make(map[uint64][]*Transaction)
	for shardID, shardPool := range txpool.RelayPool {
		for _, tx := range shardPool {
			if tx.Sender == addr {
				txTransfered = append(txTransfered, tx)
			} else {
				if _, ok := newRelayPool[shardID]; !ok {
					newRelayPool[shardID] = make([]*Transaction, 0)
				}
				newRelayPool[shardID] = append(newRelayPool[shardID], tx)
			}
		}
	}
	txpool.TxQueue = newTxQueue
	txpool.RelayPool = newRelayPool
	return txTransfered
}

// 返回某一半
func (txpool *TxPool) GetHalfAccountPartitioin(part bool) []string {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	var res []string
	for i := 0; i < len(txpool.TxQueue); i++ {
		tx := txpool.TxQueue[i]
		addr, _ := getAddrAndType(tx)
		if !txpool.AddrDivsionMap[addr] {
			res = append(res, addr)
		}
	}
	return res
}
