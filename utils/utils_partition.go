package utils

import (
	"blockEmulator/params"
	"fmt"
	"log"
	"strconv"
)

// the default method
func Addr2Shard(addr Address) int {
	last16_addr := addr[len(addr)-8:]
	num, err := strconv.ParseUint(last16_addr, 16, 64)
	if err != nil {
		log.Panic(err)
	}
	return int(num) % params.ShardNum
}

func ShowContent(target map[string]uint64) {
	idx := 0
	for key, value := range target {
		fmt.Println(key, "->", value)
		idx += 1
		if idx == 20 {
			break
		}
	}
}
