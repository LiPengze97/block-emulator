// addtional module for new consensus
package pbft_all

import (
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"blockEmulator/utils"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

// simple implementation of pbftHandleModule interface ...
// only for block request and use transaction relay
type PyramidPbftExtraHandleMod struct {
	isBshard bool
	pbftNode *PbftConsensusNode
	// pointer to pbft data
}

// propose request with different types
func (rphm *PyramidPbftExtraHandleMod) HandleinPropose() (bool, *message.Request) {
	// new blocks
	block := rphm.pbftNode.CurChain.GenerateBlock()
	r := &message.Request{
		RequestType: message.BlockRequest,
		ReqTime:     time.Now(),
	}
	r.Msg.Content = block.Encode()

	return true, r
}

// the diy operation in preprepare
func (rphm *PyramidPbftExtraHandleMod) HandleinPrePrepare(ppmsg *message.PrePrepare) bool {
	if rphm.pbftNode.CurChain.IsValidBlock(core.DecodeB(ppmsg.RequestMsg.Msg.Content)) != nil {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : not a valid block\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
		return false
	}
	rphm.pbftNode.pl.Plog.Printf("S%dN%d : the pre-prepare message is correct, putting it into the RequestPool. \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
	rphm.pbftNode.requestPool[string(ppmsg.Digest)] = ppmsg.RequestMsg
	// merge to be a prepare message
	return true
}

// the operation in prepare, and in pbft + tx relaying, this function does not need to do any.
func (rphm *PyramidPbftExtraHandleMod) HandleinPrepare(pmsg *message.Prepare) bool {
	fmt.Println("No operations are performed in Extra handle mod")
	return true
}

// the operation in commit.
func (rphm *PyramidPbftExtraHandleMod) HandleinCommit(cmsg *message.Commit) bool {
	r := rphm.pbftNode.requestPool[string(cmsg.Digest)]
	// requestType ...
	block := core.DecodeB(r.Msg.Content)
	rphm.pbftNode.pl.Plog.Printf("S%dN%d : adding the block %d...now height = %d \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, block.Header.Number, rphm.pbftNode.CurChain.CurrentBlock.Header.Number)
	rphm.pbftNode.CurChain.AddBlock(block)
	rphm.pbftNode.pl.Plog.Printf("S%dN%d : added the block %d... \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, block.Header.Number)
	rphm.pbftNode.CurChain.PrintBlockChain()

	// now try to relay txs to other shards (for main nodes)
	if rphm.pbftNode.NodeID == rphm.pbftNode.view {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : main node is trying to send relay txs at height = %d \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, block.Header.Number)
		// generate relay pool and collect txs excuted
		txExcuted := make([]*core.Transaction, 0)
		rphm.pbftNode.CurChain.Txpool.RelayPool = make(map[uint64][]*core.Transaction)
		relay1Txs := make([]*core.Transaction, 0)
		pyramidTxs := make([]*core.Transaction, 0)
		if rphm.isBshard { //Bshard特殊处理
			for _, tx := range block.Body {
				subtx1 := core.NewTransaction(tx.Sender, tx.Sender, tx.Value, tx.Nonce)
				s1sid := uint64(utils.Addr2ShardPyramid(tx.Sender))
				subtx1.BShardProcessed = true
				subtx2 := core.NewTransaction(tx.Recipient, tx.Recipient, tx.Value, tx.Nonce)
				s2sid := uint64(utils.Addr2ShardPyramid(tx.Recipient))
				subtx2.BShardProcessed = true
				rphm.pbftNode.CurChain.Txpool.AddRelayTx(subtx1, s1sid)
				rphm.pbftNode.CurChain.Txpool.AddRelayTx(subtx2, s2sid)
			}
		} else {
			for _, tx := range block.Body {
				if tx.BShardProcessed { //普通分片处理bshard处理过的交易
					pyramidTxs = append(pyramidTxs, tx)
					continue
				}
				rsid := rphm.pbftNode.CurChain.Get_PartitionMap(tx.Recipient)
				if rsid != rphm.pbftNode.ShardID {
					ntx := tx
					ntx.Relayed = true
					rphm.pbftNode.CurChain.Txpool.AddRelayTx(ntx, rsid)
					relay1Txs = append(relay1Txs, tx)
				} else {
					txExcuted = append(txExcuted, tx)
				}
			}
		}

		// send B-shard processed txs
		if rphm.isBshard {
			for sid := uint64(0); sid < rphm.pbftNode.pbftChainConfig.ShardNums; sid++ {
				if sid == rphm.pbftNode.ShardID || len(rphm.pbftNode.CurChain.Txpool.RelayPool[sid]) == 0 {
					continue
				}
				it := message.InjectTxs{
					Txs:       rphm.pbftNode.CurChain.Txpool.RelayPool[sid],
					ToShardID: sid,
					From:      rphm.pbftNode.ShardID,
				}
				itByte, err := json.Marshal(it)
				if err != nil {
					log.Panic(err)
				}
				send_msg := message.MergeMessage(message.CInject, itByte)
				go networks.TcpDial(send_msg, rphm.pbftNode.ip_nodeTable[sid][0])

				rphm.pbftNode.pl.Plog.Printf("Pyramid S%dN%d : sended inter txs to %d\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, sid)
			}
		} else {
			// send relay txs
			for sid := uint64(0); sid < rphm.pbftNode.pbftChainConfig.ShardNums; sid++ {
				if sid == rphm.pbftNode.ShardID {
					continue
				}
				relay := message.Relay{
					Txs:           rphm.pbftNode.CurChain.Txpool.RelayPool[sid],
					SenderShardID: rphm.pbftNode.ShardID,
					SenderSeq:     rphm.pbftNode.sequenceID,
				}
				rByte, err := json.Marshal(relay)
				if err != nil {
					log.Panic()
				}
				msg_send := message.MergeMessage(message.CRelay, rByte)
				go networks.TcpDial(msg_send, rphm.pbftNode.ip_nodeTable[sid][0])
				rphm.pbftNode.pl.Plog.Printf("S%dN%d : sended relay txs to %d\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, sid)
			}
		}
		rphm.pbftNode.CurChain.Txpool.ClearRelayPool()
		// send txs excuted in this block to the listener
		// add more message to measure more metrics
		bim := message.BlockInfoMsg{
			BlockBodyLength: len(block.Body),
			ExcutedTxs:      txExcuted,
			Epoch:           0,
			Relay1Txs:       relay1Txs,
			Relay1TxNum:     uint64(len(relay1Txs)),
			PyramidTxs:      pyramidTxs,
			ParamidTxNum:    uint64(len(pyramidTxs)),
			SenderShardID:   rphm.pbftNode.ShardID,
			ProposeTime:     r.ReqTime,
			CommitTime:      time.Now(),
		}
		bByte, err := json.Marshal(bim)
		if err != nil {
			log.Panic()
		}
		msg_send := message.MergeMessage(message.CBlockInfo, bByte)
		go networks.TcpDial(msg_send, rphm.pbftNode.ip_nodeTable[params.DeciderShard][0])
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : sended excuted txs\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
		rphm.pbftNode.CurChain.Txpool.GetLocked()
		rphm.pbftNode.writeCSVline([]string{strconv.Itoa(len(rphm.pbftNode.CurChain.Txpool.TxQueue)), strconv.Itoa(len(txExcuted)), strconv.Itoa(int(bim.Relay1TxNum))})
		rphm.pbftNode.CurChain.Txpool.GetUnlocked()
	}
	return true
}

func (rphm *PyramidPbftExtraHandleMod) HandleReqestforOldSeq(*message.RequestOldMessage) bool {
	fmt.Println("No operations are performed in Extra handle mod")
	return true
}

// the operation for sequential requests
func (rphm *PyramidPbftExtraHandleMod) HandleforSequentialRequest(som *message.SendOldMessage) bool {
	if int(som.SeqEndHeight-som.SeqStartHeight+1) != len(som.OldRequest) {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : the SendOldMessage message is not enough\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
	} else { // add the block into the node pbft blockchain
		for height := som.SeqStartHeight; height <= som.SeqEndHeight; height++ {
			r := som.OldRequest[height-som.SeqStartHeight]
			if r.RequestType == message.BlockRequest {
				b := core.DecodeB(r.Msg.Content)
				rphm.pbftNode.CurChain.AddBlock(b)
			}
		}
		rphm.pbftNode.sequenceID = som.SeqEndHeight + 1
		rphm.pbftNode.CurChain.PrintBlockChain()
	}
	return true
}
