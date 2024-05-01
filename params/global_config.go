package params

var (
	Block_Interval      = 5000   // generate new block interval
	MaxBlockSize_global = 2000   // the block contains the maximum number of transactions
	InjectSpeed         = 4500   // the transaction inject speed 22250 for 32
	TotalDataSize       = 250000 // the total number of txs
	BatchSize           = 36000  // supervisor read a batch of txs then send them, it should be larger than inject speed
	BrokerNum           = 10
	NodesInShard        = 3
	ShardNum            = 32
	DataWrite_path      = "./result/"       // measurement data result output path
	LogWrite_path       = "./log"           // log output path
	SupervisorAddr      = "127.0.0.1:18800" //supervisor ip address
	// FileInput           = `/mnt/e/eth_data/11000000to11999999_BlockTransaction.csv`
	FileInput         = `/users/pzl97/11000000to11999999_first_5_million.csv`
	PyramidBShardNums = 3                  // pyramid中b-shard的数量
	PyramidMappings   = map[uint64]uint64{ //看b-shard满足哪个分片到哪个分片的，是双向mapping
		0: 2,
		2: 0,
		5: 4,
		4: 5,
		7: 3,
		3: 7,
	}
	IShardBShardMappings = map[uint64]uint64{ //I-shard由哪个B-Shard负责
		0: 8,
		2: 8,
		5: 9,
		4: 9,
		7: 10,
		3: 10,
	}
)
