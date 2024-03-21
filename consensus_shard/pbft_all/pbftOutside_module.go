package pbft_all

import (
	"blockEmulator/message"
	"blockEmulator/params"
	"encoding/json"
	"log"
)

// This module used in the blockChain using transaction relaying mechanism.
// "Raw" means that the pbft only make block consensus.
type RawRelayOutsideModule struct {
	pbftNode *PbftConsensusNode
}

// msgType canbe defined in message
func (rrom *RawRelayOutsideModule) HandleMessageOutsidePBFT(msgType message.MessageType, content []byte) bool {
	switch msgType {
	case message.CRelay:
		rrom.handleRelay(content)
	case message.CInject:
		rrom.handleInjectTx(content)
	case message.CJakiroTx:
		rrom.handleJakiroInjectTx(content)
	case message.CJakiroRollupConfirm:
		rrom.handleJakiroConfirmMsg(content)
	default:
	}
	return true
}

// receive relay transaction, which is for cross shard txs
func (rrom *RawRelayOutsideModule) handleRelay(content []byte) {
	relay := new(message.Relay)
	err := json.Unmarshal(content, relay)
	if err != nil {
		log.Panic(err)
	}
	rrom.pbftNode.pl.Plog.Printf("S%dN%d : has received relay txs from shard %d, the senderSeq is %d\n", rrom.pbftNode.ShardID, rrom.pbftNode.NodeID, relay.SenderShardID, relay.SenderSeq)
	rrom.pbftNode.CurChain.Txpool.AddTxs2Pool(relay.Txs)
	rrom.pbftNode.seqMapLock.Lock()
	rrom.pbftNode.seqIDMap[relay.SenderShardID] = relay.SenderSeq
	rrom.pbftNode.seqMapLock.Unlock()
	rrom.pbftNode.pl.Plog.Printf("S%dN%d : has handled relay txs msg\n", rrom.pbftNode.ShardID, rrom.pbftNode.NodeID)
}

func (rrom *RawRelayOutsideModule) handleInjectTx(content []byte) {
	it := new(message.InjectTxs)
	err := json.Unmarshal(content, it)
	if err != nil {
		log.Panic(err)
	}
	rrom.pbftNode.CurChain.Txpool.AddTxs2Pool(it.Txs)
	rrom.pbftNode.pl.Plog.Printf("S%dN%d : has handled injected txs msg, txs: %d \n", rrom.pbftNode.ShardID, rrom.pbftNode.NodeID, len(it.Txs))
}

// 插入第二个交易池
func (rrom *RawRelayOutsideModule) handleJakiroInjectTx(content []byte) {
	it := new(message.JakiroTx)
	err := json.Unmarshal(content, it)
	if err != nil {
		log.Panic(err)
	}
	// rrom.pbftNode.pl.Plog.Printf("account:%v\n account state:%v", it.AccountAddr, it.AccountsInitialState)
	rrom.pbftNode.CurChain.Txpool2.AddTxs2Pool(it.Txs)
	rrom.pbftNode.JakiroAccountsList = it.AccountAddr
	rrom.pbftNode.CurChain.AddAccounts(it.AccountAddr, it.AccountsInitialState)
	rrom.pbftNode.pl.Plog.Printf("S%dN%d : has handled Jakiro injected txs msg, txs: %d from shard %d, now Jakiro Tx pool has %d txs\n", rrom.pbftNode.ShardID, rrom.pbftNode.NodeID, len(it.Txs), it.FromShard, len(rrom.pbftNode.CurChain.Txpool2.TxQueue))
}

// 解锁冻结的状态和交易
func (rrom *RawRelayOutsideModule) handleJakiroConfirmMsg(content []byte) {
	rrom.pbftNode.pl.Plog.Printf("S%dN%d : receive confirm the transaction rollup from shard %d\n", rrom.pbftNode.ShardID, rrom.pbftNode.NodeID, params.JakiroMapping[rrom.pbftNode.ShardID])
	rrom.pbftNode.HasSentJakiro = false
}
