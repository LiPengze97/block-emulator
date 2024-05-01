package pbft_all

import (
	"blockEmulator/message"
	"blockEmulator/utils"
	"encoding/json"
	"log"
)

// This module used in the blockChain using transaction relaying mechanism.
// "Raw" means that the pbft only make block consensus.
type PyramidOutsideModule struct {
	isBshard bool
	pbftNode *PbftConsensusNode
}

// msgType canbe defined in message
func (rrom *PyramidOutsideModule) HandleMessageOutsidePBFT(msgType message.MessageType, content []byte) bool {
	switch msgType {
	case message.CRelay:
		rrom.handleRelay(content)
	case message.CInject:
		rrom.handleInjectTx(content)
	default:
	}
	return true
}

// receive relay transaction, which is for cross shard txs
func (rrom *PyramidOutsideModule) handleRelay(content []byte) {
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

func (rrom *PyramidOutsideModule) handleInjectTx(content []byte) {
	it := new(message.InjectTxs)
	err := json.Unmarshal(content, it)
	if err != nil {
		log.Panic(err)
	}
	rrom.pbftNode.CurChain.Txpool.AddTxs2Pool(it.Txs)
	if rrom.isBshard {
		for _, tx := range it.Txs {
			sid := utils.Addr2ShardPyramid(tx.Sender)
			rid := utils.Addr2ShardPyramid(tx.Recipient)
			rrom.pbftNode.pl.Plog.Printf("tx : %v to %v \n", sid, rid)
		}
	} else {
		if it.From != 0 {
			sid := utils.Addr2ShardPyramid(it.Txs[0].Sender)
			rid := utils.Addr2ShardPyramid(it.Txs[0].Recipient)
			rrom.pbftNode.pl.Plog.Printf("received %d txs from B-shard %v, 1st tx : %v to %v \n", len(it.Txs), it.From, sid, rid)
		}
	}

	rrom.pbftNode.pl.Plog.Printf("S%dN%d : has handled injected txs msg, txs: %d \n", rrom.pbftNode.ShardID, rrom.pbftNode.NodeID, len(it.Txs))
}
