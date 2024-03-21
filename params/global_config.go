package params

var (
	Block_Interval      = 5000  // generate new block interval
	MaxBlockSize_global = 1000  // the block contains the maximum number of transactions
	InjectSpeed         = 2250  // the transaction inject speed 22250 for 32
	TotalDataSize       = 50000 // the total number of txs
	BatchSize           = 18000 // supervisor read a batch of txs then send them, it should be larger than inject speed
	BrokerNum           = 10
	NodesInShard        = 3
	ShardNum            = 32
	DataWrite_path      = "./result/"                                               // measurement data result output path
	LogWrite_path       = "./log"                                                   // log output path
	SupervisorAddr      = "127.0.0.1:18800"                                         //supervisor ip address
	FileInput           = `/mnt/e/eth_data/11000000to11999999_BlockTransaction.csv` //the raw BlockTransaction data path
	// FileInput     = `/users/pzl97/11000000to11999999_first_5million_rows.csv` //the raw BlockTransaction data path

	JakiroMapping = map[uint64]uint64{
		0: 2,
		1: 3,
	}
	// JakiroMapping = map[uint64]uint64{
	// 	8: 2,
	// 	15: 14,
	// 	30: 0,
	// 	5: 1,
	// 	31: 20,
	// 	13: 23,
	// 	21: 29,
	// 	6: 25,
	// 	3: 4,
	// 	16: 7,
	// }
	JakiroThreshold = 4 // 交易池里面的交易超过多少就发jakiro消息
)
