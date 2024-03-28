package params

var (
	Block_Interval      = 5000   // generate new block interval
	MaxBlockSize_global = 2000   // the block contains the maximum number of transactions
	InjectSpeed         = 9000   // the transaction inject speed 22250 for 32
	TotalDataSize       = 500000 // the total number of txs
	BatchSize           = 18000  // supervisor read a batch of txs then send them, it should be larger than inject speed
	BrokerNum           = 10
	NodesInShard        = 3
	ShardNum            = 32
	DataWrite_path      = "./result/"       // measurement data result output path
	LogWrite_path       = "./log"           // log output path
	SupervisorAddr      = "127.0.0.1:18800" //supervisor ip address
	// FileInput           = `/mnt/e/eth_data/11000000to11999999_BlockTransaction.csv`
	FileInput = `/mnt/e/eth_data/16000000to16249999_BlockTransaction.csv`
	// FileInput = `/mnt/e/eth_data/17000000to17249999_BlockTransaction.csv`

	// default
	JakiroMapping = map[uint64]uint64{
		123: 345,
	}

	// 4
	// JakiroMapping = map[uint64]uint64{
	// 	0: 2,
	// 	1: 3,
	// }

	// // 8
	// JakiroMapping = map[uint64]uint64{
	// 	0: 2,
	// 	5: 4,
	// 	7: 3,
	// }

	// 16
	// JakiroMapping = map[uint64]uint64{
	// 	15: 2,
	// 	8:  4,
	// 	5:  7,
	// 	14: 11,
	// 	13: 12,
	// 	3:  1,
	// }

	// 32
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
