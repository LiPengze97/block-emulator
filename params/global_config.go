package params

import "fmt"

var (
	Block_Interval       = 3000   // generate new block interval
	MaxBlockSize_global  = 2000   // the block contains the maximum number of transactions
	InjectSpeed          = 1000   // the transaction inject speed
	TotalDataSize        = 100000 // the total number of txs
	BatchSize            = 4000   // supervisor read a batch of txs then send them, it should be larger than inject speed
	BrokerNum            = 10
	NodesInShard         = 4
	ShardNum             = 4
	DataWrite_path       = "./result/"                                                                               // measurement data result output path
	LogWrite_path        = "./log"                                                                                   // log output path
	SupervisorAddr       = "127.0.0.1:18800"                                                                         //supervisor ip address
	FileInput            = `/mnt/e/eth_data/2000000to2999999_BlockTransaction/2000000to2999999_BlockTransaction.csv` //the raw BlockTransaction data path
	AllocationInput      = fmt.Sprintf("/mnt/d/eth_data/SpringWWW/dataAndPlacement/placement_addresses_%v.json", YearOfData)
	BaseDataPath         = fmt.Sprintf("/mnt/d/eth_data/SpringWWW/dataAndPlacement/%v.csv", YearOfData)
	YearOfData           = 2023
	OtherAllocationInput = fmt.Sprintf("/mnt/d/eth_data/SpringWWW/dataAndPlacement/%v-2015.json", AllocaionMethod)
	AllocaionMethod      = "Spring" // Spring Monoxide Random Skychain ShardScheduler
)
