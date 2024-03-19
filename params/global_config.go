package params

var (
	Block_Interval      = 5000   // generate new block interval
	MaxBlockSize_global = 2000   // the block contains the maximum number of transactions
	InjectSpeed         = 4500   // the transaction inject speed 22250 for 32
	TotalDataSize       = 250000 // the total number of txs
	BatchSize           = 8000   // supervisor read a batch of txs then send them, it should be larger than inject speed
	BrokerNum           = 10
	NodesInShard        = 3
	ShardNum            = 32
	DataWrite_path      = "./result/"       // measurement data result output path
	LogWrite_path       = "./log"           // log output path
	SupervisorAddr      = "127.0.0.1:18800" //supervisor ip address
	// FileInput           = `/mnt/e/eth_data/2000000to2999999_BlockTransaction/2000000to2999999_BlockTransaction.csv` //the raw BlockTransaction data path
	FileInput     = `/mnt/e/eth_data/11000000to11999999_BlockTransaction.csv` //the raw BlockTransaction data path
	JakiroMapping = map[uint64]uint64{
		100: 2,
		111: 3,
	}
	JakiroThreshold = 4 // 交易池里面的交易超过多少就发jakiro消息
)
