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
type RawRelayPbftExtraHandleMod struct {
	pbftNode *PbftConsensusNode
	// pointer to pbft data
}

// propose request with different types
func (rphm *RawRelayPbftExtraHandleMod) HandleinPropose() (bool, *message.Request) {
	// new blocks

	r := &message.Request{
		RequestType:     message.BlockRequest,
		CanSendJakiroTx: rphm.pbftNode.CanSendJakiroTx([]string{}),
	}

	if r.CanSendJakiroTx {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : propose jakiro message\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
		rphm.pbftNode.HasSentJakiro = true
		// 如果可以发送JakiroTx，直接准备好要发送的地址
		r.AccountCandidate = rphm.pbftNode.CurChain.Txpool.GetHalfAccountPartitioin(utils.Part2)
	}

	jakiroBlock := rphm.pbftNode.CurChain.GenerateJakiroBlock()

	if len(jakiroBlock.Body2) > 0 {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : propose jakiro block with %d txs from jakiro txpool\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, len(jakiroBlock.Body2))
	}

	r.Msg.Content = jakiroBlock.Encode()

	r.ReqTime = time.Now()

	return true, r
}

// the diy operation in preprepare
func (rphm *RawRelayPbftExtraHandleMod) HandleinPrePrepare(ppmsg *message.PrePrepare) bool {

	if rphm.pbftNode.CurChain.IsValidJakiroBlock(core.DecodeJB(ppmsg.RequestMsg.Msg.Content)) != nil {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : not a valid block\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
		return false
	}
	rphm.pbftNode.pl.Plog.Printf("S%dN%d : the pre-prepare message is correct, putting it into the RequestPool. \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
	rphm.pbftNode.requestPool[string(ppmsg.Digest)] = ppmsg.RequestMsg
	// merge to be a prepare message
	return true
}

// the operation in prepare, and in pbft + tx relaying, this function does not need to do any.
func (rphm *RawRelayPbftExtraHandleMod) HandleinPrepare(pmsg *message.Prepare) bool {
	fmt.Println("No operations are performed in Extra handle mod")
	return true
}

// the operation in commit.
func (rphm *RawRelayPbftExtraHandleMod) HandleinCommit(cmsg *message.Commit) bool {
	r := rphm.pbftNode.requestPool[string(cmsg.Digest)]
	// requestType ...
	block := core.DecodeJB(r.Msg.Content)
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
		// Jakiro
		jakiroTx := make([]*core.Transaction, 0)
		// 先处理Jakiro交易池里面的交易，这样才可以把跨链交易也变成链内交易，这部分逻辑是给负载低的链的
		if block.Body2 != nil {
			for _, tx := range block.Body2 {
				rsid := rphm.pbftNode.CurChain.Get_PartitionMap(tx.Recipient)
				ssid := rphm.pbftNode.CurChain.Get_PartitionMap(tx.Sender)
				if tx.Relayed {
					jakiroTx = append(jakiroTx, tx)
				} else if rsid != rphm.pbftNode.ShardID && rsid != ssid {
					ntx := tx
					ntx.Relayed = true
					rphm.pbftNode.CurChain.Txpool.AddRelayTx(ntx, rsid)
					relay1Txs = append(relay1Txs, tx)
				} else {
					jakiroTx = append(jakiroTx, tx)
				}
			}
		}

		for _, tx := range block.Body {
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
		rphm.pbftNode.CurChain.Txpool.ClearRelayPool()

		// 负载高的发送交易给负载低的，只需要主节点做
		if rphm.pbftNode.NodeID == 0 && rphm.pbftNode.CanSendJakiroTxs {
			targetSid := params.JakiroMapping[rphm.pbftNode.ShardID]
			addrToBeSent := make([]string, 0)
			for k, v := range rphm.pbftNode.CurChain.Txpool.AddrDivsionMap {
				if !v {
					addrToBeSent = append(addrToBeSent, k)
				}
			}
			jakiro := message.JakiroTx{
				Txs:                  rphm.pbftNode.CurChain.Txpool.SplitTxs(),
				AccountsInitialState: rphm.pbftNode.CurChain.FetchAccounts(addrToBeSent),
				AccountAddr:          addrToBeSent,
				FromShard:            rphm.pbftNode.ShardID,
			}

			jByte, err := json.Marshal(jakiro)
			if err != nil {
				log.Panic()
			}
			msg_send := message.MergeMessage(message.CJakiroTx, jByte)
			go networks.TcpDial(msg_send, rphm.pbftNode.ip_nodeTable[targetSid][0])
			rphm.pbftNode.pl.Plog.Printf("S%dN%d : sended jakiro %d txs to %d\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, len(jakiro.Txs), targetSid)
			rphm.pbftNode.CanSendJakiroTxs = false

		}

		// 2024-02-21TODO只需要在confirm的时候，加最后的验证部分就行，账户状态。先在收到的时候，保存对应的用户地址，保留一系列的区块哈希，然后得出结论就行。

		// 负载低的给负载高的发送确认消息，只需要主节点做
		if rphm.pbftNode.NodeID == 0 && !rphm.pbftNode.CurChain.IsHighLoadedChain {
			// 记录当前区块的哈希
			if len(block.Body2) > 0 {
				rphm.pbftNode.JakiroHeaders = append(rphm.pbftNode.JakiroHeaders, block.Header.PrintBlockHeader())
			}
			// 如果处理完了，再发一个完整的消息给源链
			if len(block.Body2) > 0 && len(rphm.pbftNode.CurChain.Txpool2.TxQueue) == 0 {
				jakiroConfirm := message.JakiroConfirm{
					FromShard:      rphm.pbftNode.ShardID,
					ConfirmHeaders: rphm.pbftNode.JakiroHeaders,
					AccountStates:  rphm.pbftNode.CurChain.FetchAccounts(rphm.pbftNode.JakiroAccountsList),
				}
				// 清空Header
				rphm.pbftNode.JakiroHeaders = []string{}
				jByte, err := json.Marshal(jakiroConfirm)
				if err != nil {
					log.Panic()
				}
				msg_send := message.MergeMessage(message.CJakiroRollupConfirm, jByte)
				go networks.TcpDial(msg_send, rphm.pbftNode.ip_nodeTable[rphm.pbftNode.CurChain.JakiroFrom][0])
				rphm.pbftNode.pl.Plog.Printf("S%dN%d : done processing sended jakiro confirm to %d\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, rphm.pbftNode.CurChain.JakiroFrom)
			}

		}

		// send txs excuted in this block to the listener

		// add more message to measure more metrics
		bim := message.BlockInfoMsg{
			BlockBodyLength:      len(block.Body),
			ExcutedTxs:           txExcuted,
			Epoch:                0,
			Relay1Txs:            relay1Txs,
			Relay1TxNum:          uint64(len(relay1Txs)),
			JakiroTxs:            jakiroTx,
			ProcessedJakiroTxNum: uint64(len(jakiroTx)),
			SenderShardID:        rphm.pbftNode.ShardID,
			ProposeTime:          r.ReqTime,
			CommitTime:           time.Now(),
		}
		bByte, err := json.Marshal(bim)
		if err != nil {
			log.Panic()
		}
		msg_send := message.MergeMessage(message.CBlockInfo, bByte)
		go networks.TcpDial(msg_send, rphm.pbftNode.ip_nodeTable[params.DeciderShard][0])
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : sended excuted txs\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
		rphm.pbftNode.CurChain.Txpool.GetLocked()
		rphm.pbftNode.writeCSVline([]string{strconv.Itoa(len(rphm.pbftNode.CurChain.Txpool.TxQueue)), strconv.Itoa(len(txExcuted)), strconv.Itoa(int(bim.Relay1TxNum)), strconv.Itoa(int(bim.ProcessedJakiroTxNum))})
		rphm.pbftNode.CurChain.Txpool.GetUnlocked()
	}
	return true
}

func (rphm *RawRelayPbftExtraHandleMod) HandleReqestforOldSeq(*message.RequestOldMessage) bool {
	fmt.Println("No operations are performed in Extra handle mod")
	return true
}

// the operation for sequential requests
func (rphm *RawRelayPbftExtraHandleMod) HandleforSequentialRequest(som *message.SendOldMessage) bool {
	if int(som.SeqEndHeight-som.SeqStartHeight+1) != len(som.OldRequest) {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : the SendOldMessage message is not enough\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
	} else { // add the block into the node pbft blockchain
		for height := som.SeqStartHeight; height <= som.SeqEndHeight; height++ {
			r := som.OldRequest[height-som.SeqStartHeight]
			if r.RequestType == message.BlockRequest {
				b := core.DecodeJB(r.Msg.Content)
				rphm.pbftNode.CurChain.AddBlock(b)
			}
		}
		rphm.pbftNode.sequenceID = som.SeqEndHeight + 1
		rphm.pbftNode.CurChain.PrintBlockChain()
	}
	return true
}
