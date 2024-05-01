package committee

import (
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"blockEmulator/supervisor/signal"
	"blockEmulator/supervisor/supervisor_log"
	"blockEmulator/utils"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"os"
	"time"
)

type PyramidCommitteeModule struct {
	csvPath      string
	dataTotalNum int
	nowDataNum   int
	batchDataNum int
	IpNodeTable  map[uint64]map[uint64]string
	sl           *supervisor_log.SupervisorLog
	Ss           *signal.StopSignal // to control the stop message sending
}

func NewPyramidCommitteeModule(Ip_nodeTable map[uint64]map[uint64]string, Ss *signal.StopSignal, slog *supervisor_log.SupervisorLog, csvFilePath string, dataNum, batchNum int) *PyramidCommitteeModule {
	return &PyramidCommitteeModule{
		csvPath:      csvFilePath,
		dataTotalNum: dataNum,
		batchDataNum: batchNum,
		nowDataNum:   0,
		IpNodeTable:  Ip_nodeTable,
		Ss:           Ss,
		sl:           slog,
	}
}

func (pthm *PyramidCommitteeModule) HandleOtherMessage([]byte) {}

func (pthm *PyramidCommitteeModule) txSending(txlist []*core.Transaction) {
	// the txs will be sent
	sendToShard := make(map[uint64][]*core.Transaction)

	for idx := 0; idx <= len(txlist); idx++ {
		if idx > 0 && (idx%params.InjectSpeed == 0 || idx == len(txlist)) {
			// send to shard
			for sid := uint64(0); sid < uint64(params.ShardNum); sid++ {
				it := message.InjectTxs{
					Txs:       sendToShard[sid],
					ToShardID: sid,
					From:      0,
				}
				itByte, err := json.Marshal(it)
				if err != nil {
					log.Panic(err)
				}
				send_msg := message.MergeMessage(message.CInject, itByte)
				go networks.TcpDial(send_msg, pthm.IpNodeTable[sid][0])
			}
			sendToShard = make(map[uint64][]*core.Transaction)
			time.Sleep(time.Second)
		}
		if idx == len(txlist) {
			break
		}
		tx := txlist[idx]
		sendersid := uint64(utils.Addr2ShardPyramid(tx.Sender))
		receiversid := uint64(utils.Addr2ShardPyramid(tx.Recipient))
		targetsid, hasBShardMapping := params.PyramidMappings[sendersid]
		if hasBShardMapping && targetsid == receiversid {
			bshardID := params.IShardBShardMappings[sendersid]
			sendToShard[bshardID] = append(sendToShard[bshardID], tx)
		} else {
			sendToShard[sendersid] = append(sendToShard[sendersid], tx)
		}

	}
}

// read transactions, the Number of the transactions is - batchDataNum
func (pthm *PyramidCommitteeModule) MsgSendingControl() {
	txfile, err := os.Open(pthm.csvPath)
	if err != nil {
		log.Panic(err)
	}
	defer txfile.Close()
	reader := csv.NewReader(txfile)
	txlist := make([]*core.Transaction, 0) // save the txs in this epoch (round)

	for {
		data, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Panic(err)
		}
		if tx, ok := data2tx(data, uint64(pthm.nowDataNum)); ok {
			txlist = append(txlist, tx)
			pthm.nowDataNum++
		}

		// re-shard condition, enough edges
		if len(txlist) == int(pthm.batchDataNum) || pthm.nowDataNum == pthm.dataTotalNum {
			pthm.txSending(txlist)
			// reset the variants about tx sending
			txlist = make([]*core.Transaction, 0)
			pthm.Ss.StopGap_Reset()
		}

		if pthm.nowDataNum == pthm.dataTotalNum {
			break
		}
	}
}

// no operation here
func (pthm *PyramidCommitteeModule) HandleBlockInfo(b *message.BlockInfoMsg) {
	pthm.sl.Slog.Printf("received from shard %d in epoch %d.\n", b.SenderShardID, b.Epoch)
}
