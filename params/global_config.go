package params

var (
	Block_Interval      = 5000   // generate new block interval
	MaxBlockSize_global = 2000   // the block contains the maximum number of transactions
	InjectSpeed         = 18000  // the transaction inject speed 22250 for 32
	TotalDataSize       = 100000 // the total number of txs
	BatchSize           = 2250   // supervisor read a batch of txs then send them, it should be larger than inject speed
	BrokerNum           = 10
	NodesInShard        = 3
	ShardNum            = 32
	DataWrite_path      = "./result/"       // measurement data result output path
	LogWrite_path       = "./log"           // log output path
	SupervisorAddr      = "127.0.0.1:18800" //supervisor ip address
	// FileInput           = `/mnt/e/eth_data/11000000to11999999_BlockTransaction.csv`
	FileInput         = `/mnt/e/eth_data/16000000to16249999_BlockTransaction.csv`
	PyramidBShardNums = 1                  // pyramid中b-shard的数量
	PyramidMappings   = map[uint64]uint64{ //看b-shard满足哪个分片到哪个分片的，是双向mapping
		0: 1,
		1: 0,
	}
	IShardBShardMappings = map[uint64]uint64{ //I-shard由哪个B-Shard负责
		0: 4,
		1: 4,
	}
)
