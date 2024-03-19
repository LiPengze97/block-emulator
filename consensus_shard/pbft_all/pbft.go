// The pbft consensus process

package pbft_all

import (
	"blockEmulator/chain"
	"blockEmulator/consensus_shard/pbft_all/pbft_log"
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"blockEmulator/shard"
	"bufio"
	"io"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
)

type PbftConsensusNode struct {
	// the local config about pbft
	RunningNode *shard.Node // the node information
	ShardID     uint64      // denote the ID of the shard (or pbft), only one pbft consensus in a shard
	NodeID      uint64      // denote the ID of the node in the pbft (shard)

	// the data structure for blockchain
	CurChain *chain.BlockChain // all node in the shard maintain the same blockchain
	db       ethdb.Database    // to save the mpt

	// the global config about pbft
	pbftChainConfig *params.ChainConfig          // the chain config in this pbft
	ip_nodeTable    map[uint64]map[uint64]string // denote the ip of the specific node
	node_nums       uint64                       // the number of nodes in this pfbt, denoted by N
	malicious_nums  uint64                       // f, 3f + 1 = N
	view            uint64                       // denote the view of this pbft, the main node can be inferred from this variant

	// the control message and message checking utils in pbft
	sequenceID         uint64                          // the message sequence id of the pbft
	stop               bool                            // send stop signal
	pStop              chan uint64                     // channle for stopping consensus
	requestPool        map[string]*message.Request     // RequestHash to Request
	cntPrepareConfirm  map[string]map[*shard.Node]bool // count the prepare confirm message, [messageHash][Node]bool
	cntPrepareJakiro   map[string]map[*shard.Node]bool // 计算Jakiro消息投了什么票 [messageHash][Node]bool prepare里面
	cntCommitConfirm   map[string]map[*shard.Node]bool // count the commit confirm message, [messageHash][Node]bool
	cntCommitJakiro    map[string]map[*shard.Node]bool // 计算Jakiro消息投了什么票 [messageHash][Node]bool commit里面
	JakiroHeaders      []string                        // 低负载链上节点处理来自高负载的交易的区块头列表
	JakiroAccountsList []string                        //低负载链上节点处理来自高负载的交易的所有账户的列表
	isCommitBordcast   map[string]bool                 // denote whether the commit is broadcast
	isReply            map[string]bool                 // denote whether the message is reply
	height2Digest      map[uint64]string               // sequence (block height) -> request, fast read

	// locks about pbft
	sequenceLock sync.Mutex // the lock of sequence
	lock         sync.Mutex // lock the stage
	askForLock   sync.Mutex // lock for asking for a serise of requests
	stopLock     sync.Mutex // lock the stop varient

	// seqID of other Shards, to synchronize
	seqIDMap   map[uint64]uint64
	seqMapLock sync.Mutex

	// logger
	pl *pbft_log.PbftLog
	// tcp control
	tcpln       net.Listener
	tcpPoolLock sync.Mutex

	// to handle the message in the pbft
	ihm ExtraOpInConsensus

	// to handle the message outside of pbft
	ohm OpInterShards

	// 负载重的分片有没有已经发送的Jakiro交易
	HasSentJakiro bool

	// Commit消息判断是否可以把自己交易池里的交易发给另一个链了
	CanSendJakiroTxs bool

	// 负载重的分片保存要发送的状态
	AccountStatesToBeSent []*core.AccountState
}

// generate a pbft consensus for a node
func NewPbftNode(shardID, nodeID uint64, pcc *params.ChainConfig, messageHandleType string) *PbftConsensusNode {
	p := new(PbftConsensusNode)
	p.ip_nodeTable = params.IPmap_nodeTable
	p.node_nums = pcc.Nodes_perShard
	p.ShardID = shardID
	p.NodeID = nodeID
	p.pbftChainConfig = pcc
	p.HasSentJakiro = false
	p.CanSendJakiroTxs = false
	fp := "./record/ldb/s" + strconv.FormatUint(shardID, 10) + "/n" + strconv.FormatUint(nodeID, 10)
	var err error
	p.db, err = rawdb.NewLevelDBDatabase(fp, 0, 1, "accountState", false)
	if err != nil {
		log.Panic(err)
	}

	p.CurChain, err = chain.NewBlockChain(pcc, p.db)

	if err != nil {
		log.Panic("cannot new a blockchain")
	}

	if _, ok := params.JakiroMapping[shardID]; ok {
		p.CurChain.IsHighLoadedChain = true
		p.CurChain.Txpool.IsJakiroTxSourcePool = true
	}

	for k, v := range params.JakiroMapping {
		if v == p.ShardID {
			p.CurChain.JakiroFrom = k
			break
		}
	}

	p.RunningNode = &shard.Node{
		NodeID:  nodeID,
		ShardID: shardID,
		IPaddr:  p.ip_nodeTable[shardID][nodeID],
	}

	p.stop = false
	p.sequenceID = p.CurChain.CurrentBlock.Header.Number + 1
	p.pStop = make(chan uint64)
	p.requestPool = make(map[string]*message.Request)
	p.cntPrepareConfirm = make(map[string]map[*shard.Node]bool)
	p.cntPrepareJakiro = make(map[string]map[*shard.Node]bool)
	p.cntCommitConfirm = make(map[string]map[*shard.Node]bool)
	p.cntCommitJakiro = make(map[string]map[*shard.Node]bool)
	p.isCommitBordcast = make(map[string]bool)
	p.isReply = make(map[string]bool)
	p.height2Digest = make(map[uint64]string)
	p.malicious_nums = (p.node_nums - 1) / 3
	p.view = 0

	p.seqIDMap = make(map[uint64]uint64)

	p.pl = pbft_log.NewPbftLog(shardID, nodeID)

	// if p.CurChain.IsHighLoadedChain {
	// 	p.pl.Plog.Printf("S%dN%d enables Jakiro. \n", p.ShardID, p.NodeID)
	// }

	// choose how to handle the messages in pbft or beyond pbft
	switch string(messageHandleType) {
	default:
		p.ihm = &RawRelayPbftExtraHandleMod{
			pbftNode: p,
		}
		p.ohm = &RawRelayOutsideModule{
			pbftNode: p,
		}
	}

	return p
}

// 根据自己的链的情况和交易池里面的负载，判断是否可以发送Jakiro交易
func (p *PbftConsensusNode) CanSendJakiroTx(accountCandidates []string) bool {
	if p.NodeID != 0 {
		// 由于BlockEmulator里面的从节点并没有真实收到交易，所以这里先注释掉，直接return true
		/*
			candidateSum := 0
			half := p.CurChain.Txpool.GetTxQueueLen() / 2
			for _, account := range accountCandidates {
				candidateSum += int(p.CurChain.Txpool.TxNumMap[account].Sum)
			}
			// 容忍一定的误差，这里设置的是20%的误差
			if half*4/5 <= 1.0*candidateSum {
				return true
			} else {
				return false
			}
		*/
		return true
	}

	if p.CurChain.IsHighLoadedChain && int(len(p.CurChain.Txpool.TxQueue))/params.MaxBlockSize_global >= params.JakiroThreshold && !p.HasSentJakiro {
		return true
	} else {
		return false
	}
}

// handle the raw message, send it to corresponded interfaces
func (p *PbftConsensusNode) handleMessage(msg []byte) {
	msgType, content := message.SplitMessage(msg)
	switch msgType {
	// pbft inside message type
	case message.CPrePrepare:
		p.handlePrePrepare(content)
	case message.CPrepare:
		p.handlePrepare(content)
	case message.CCommit:
		p.handleCommit(content)
	case message.CRequestOldrequest:
		p.handleRequestOldSeq(content)
	case message.CSendOldrequest:
		p.handleSendOldSeq(content)
	case message.CStop:
		p.WaitToStop()

	// handle the message from outside
	default:
		p.ohm.HandleMessageOutsidePBFT(msgType, content)
	}
}

func (p *PbftConsensusNode) handleClientRequest(con net.Conn) {
	defer con.Close()
	clientReader := bufio.NewReader(con)
	for {
		clientRequest, err := clientReader.ReadBytes('\n')
		if p.getStopSignal() {
			return
		}
		switch err {
		case nil:
			p.tcpPoolLock.Lock()
			p.handleMessage(clientRequest)
			p.tcpPoolLock.Unlock()
		case io.EOF:
			log.Println("client closed the connection by terminating the process")
			return
		default:
			log.Printf("error: %v\n", err)
			return
		}
	}
}

func (p *PbftConsensusNode) TcpListen() {
	ln, err := net.Listen("tcp", p.RunningNode.IPaddr)
	p.tcpln = ln
	if err != nil {
		log.Panic(err)
	}
	for {
		conn, err := p.tcpln.Accept()
		if err != nil {
			return
		}
		go p.handleClientRequest(conn)
	}
}

// listen to the request
func (p *PbftConsensusNode) OldTcpListen() {
	ipaddr, err := net.ResolveTCPAddr("tcp", p.RunningNode.IPaddr)
	if err != nil {
		log.Panic(err)
	}
	ln, err := net.ListenTCP("tcp", ipaddr)
	p.tcpln = ln
	if err != nil {
		log.Panic(err)
	}
	p.pl.Plog.Printf("S%dN%d begins listening：%s\n", p.ShardID, p.NodeID, p.RunningNode.IPaddr)

	for {
		if p.getStopSignal() {
			p.closePbft()
			return
		}
		conn, err := p.tcpln.Accept()
		if err != nil {
			log.Panic(err)
		}
		b, err := io.ReadAll(conn)
		if err != nil {
			log.Panic(err)
		}
		p.handleMessage(b)
		conn.(*net.TCPConn).SetLinger(0)
		defer conn.Close()
	}
}

// when received stop
func (p *PbftConsensusNode) WaitToStop() {
	p.pl.Plog.Println("handling stop message")
	p.stopLock.Lock()
	p.stop = true
	p.stopLock.Unlock()
	if p.NodeID == p.view {
		p.pStop <- 1
	}
	networks.CloseAllConnInPool()
	p.tcpln.Close()
	p.closePbft()
	p.pl.Plog.Println("handled stop message")
}

func (p *PbftConsensusNode) getStopSignal() bool {
	p.stopLock.Lock()
	defer p.stopLock.Unlock()
	return p.stop
}

// close the pbft
func (p *PbftConsensusNode) closePbft() {
	p.CurChain.CloseBlockChain()
}
