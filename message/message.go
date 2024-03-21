package message

import (
	"blockEmulator/core"
	"blockEmulator/shard"
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

var prefixMSGtypeLen = 30

type MessageType string
type RequestType string

const (
	CPrePrepare        MessageType = "preprepare"
	CPrepare           MessageType = "prepare"
	CCommit            MessageType = "commit"
	CRequestOldrequest MessageType = "requestOldrequest"
	CSendOldrequest    MessageType = "sendOldrequest"
	CStop              MessageType = "stop"

	CRelay  MessageType = "relay"
	CInject MessageType = "inject"

	CBlockInfo MessageType = "BlockInfo"
	CSeqIDinfo MessageType = "SequenceID"

	CJakiroTx            MessageType = "jakirotx"
	CJakiroByteTx        MessageType = "jakiroByteTx"
	CJakiroRollupConfirm MessageType = "jakirorollup"
)

var (
	BlockRequest RequestType = "Block"
	// add more types
	// ...
)

type RawMessage struct {
	Content []byte // the content of raw message, txs and blocks (most cases) included
}

type Request struct {
	RequestType      RequestType
	Msg              RawMessage // request message
	ReqTime          time.Time  // request time
	CanSendJakiroTx  bool
	AccountCandidate []string //如果CanSendJakiroTx为true，这里需要添加相应内容用于验证
}

type PrePrepare struct {
	RequestMsg *Request // the request message should be pre-prepared
	Digest     []byte   // the digest of this request, which is the only identifier
	SeqID      uint64
}

type Prepare struct {
	Digest          []byte // To identify which request is prepared by this node
	SeqID           uint64
	SenderNode      *shard.Node // To identify who send this message
	CanSendJakiroTx bool
}

type Commit struct {
	Digest          []byte // To identify which request is prepared by this node
	SeqID           uint64
	SenderNode      *shard.Node // To identify who send this message
	CanSendJakiroTx bool
}

type Reply struct {
	MessageID  uint64
	SenderNode *shard.Node
	Result     bool
}

type RequestOldMessage struct {
	SeqStartHeight uint64
	SeqEndHeight   uint64
	ServerNode     *shard.Node // send this request to the server node
	SenderNode     *shard.Node
}

type SendOldMessage struct {
	SeqStartHeight uint64
	SeqEndHeight   uint64
	OldRequest     []*Request
	SenderNode     *shard.Node
}

type InjectTxs struct {
	Txs       []*core.Transaction
	ToShardID uint64
}

type JakiroTx struct {
	Txs                  []*core.Transaction
	AccountsInitialState []*core.AccountState
	AccountAddr          []string
	FromShard            uint64
}

func (jtx *JakiroTx) Encode() []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(jtx)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func DecodeJakiroTxMsg(content []byte) *JakiroTx {
	var jtx JakiroTx

	decoder := gob.NewDecoder(bytes.NewReader(content))
	err := decoder.Decode(&jtx)
	if err != nil {
		log.Panic(err)
	}

	return &jtx
}

type JakiroConfirm struct {
	ConfirmHeaders []string
	FromShard      uint64
	AccountStates  []*core.AccountState
}

type BlockInfoMsg struct {
	BlockBodyLength int
	ExcutedTxs      []*core.Transaction // txs which are excuted completely
	Epoch           int

	ProposeTime   time.Time // record the propose time of this block (txs)
	CommitTime    time.Time // record the commit time of this block (txs)
	SenderShardID uint64

	// for transaction relay
	Relay1TxNum uint64              // the number of cross shard txs
	Relay1Txs   []*core.Transaction // cross transactions in chain first time

	// for broker
	Broker1TxNum uint64              // the number of broker 1
	Broker1Txs   []*core.Transaction // cross transactions at first time by broker
	Broker2TxNum uint64              // the number of broker 2
	Broker2Txs   []*core.Transaction // cross transactions at second time by broker

	// for Jakiro
	ProcessedJakiroTxNum uint64              // 负载轻的链处理的交易
	JakiroTxs            []*core.Transaction // 负载轻的链处理的交易
}

type SeqIDinfo struct {
	SenderShardID uint64
	SenderSeq     uint64
}

func MergeMessage(msgType MessageType, content []byte) []byte {
	b := make([]byte, prefixMSGtypeLen)
	for i, v := range []byte(msgType) {
		b[i] = v
	}
	merge := append(b, content...)
	return merge
}

func SplitMessage(message []byte) (MessageType, []byte) {
	msgTypeBytes := message[:prefixMSGtypeLen]
	msgType_pruned := make([]byte, 0)
	for _, v := range msgTypeBytes {
		if v != byte(0) {
			msgType_pruned = append(msgType_pruned, v)
		}
	}
	msgType := string(msgType_pruned)
	content := message[prefixMSGtypeLen:]
	return MessageType(msgType), content
}
