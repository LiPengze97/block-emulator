package main

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	// 创建一个空的状态数据库
	db := state.NewDatabase(nil)

	// 准备一些测试数据
	var addresses []common.Address
	for i := 0; i < 1000; i++ {
		address := crypto.Keccak256Hash([]byte(fmt.Sprintf("address%d", i)))
		addresses = append(addresses, common.BytesToAddress(address.Bytes()))
	}

	// 测试写入速度
	startTime := time.Now()
	for _, address := range addresses {
		// 为每个地址创建一个新的状态对象
		stateObject, _ := db.CreateAccount(address)

		// 设置一些示例状态数据
		stateObject.SetBalance(big.NewInt(100))
		stateObject.SetNonce(uint64(1))
		stateObject.SetCode(crypto.Keccak256Hash([]byte("code")), []byte("code"))
		stateObject.SetState(crypto.Keccak256Hash([]byte("key")), crypto.Keccak256Hash([]byte("value")))

		// 将更改写入数据库
		db.Commit(stateObject)
	}
	writeDuration := time.Since(startTime)

	// 测试读取速度
	startTime = time.Now()
	for _, address := range addresses {
		// 从数据库中获取状态对象
		stateObject, _ := db.GetAccount(address)

		// 读取状态数据
		_ = stateObject.Balance()
		_ = stateObject.Nonce()
		_ = stateObject.CodeHash()
		_ = stateObject.GetState(crypto.Keccak256Hash([]byte("key")))
	}
	readDuration := time.Since(startTime)

	// 打印结果
	fmt.Printf("写入时间: %v\n", writeDuration)
	fmt.Printf("读取时间: %v\n", readDuration)
}
