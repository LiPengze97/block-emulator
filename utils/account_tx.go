package utils

// 储存每个地址里面对应的交易数量
type AccountTxNum struct {
	IntraTx  uint32
	Relay1Tx uint32
	Relay2Tx uint32
	Sum      uint32
}

func NewAccountTxNum() *AccountTxNum {
	return &AccountTxNum{
		IntraTx:  0,
		Relay1Tx: 0,
		Relay2Tx: 0,
		Sum:      0,
	}
}
