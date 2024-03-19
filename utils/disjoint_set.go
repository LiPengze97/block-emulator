package utils

import "fmt"

type DisjointSet struct {
	parent map[string]string
	rank   map[string]int
}

func NewDisjointSet() *DisjointSet {
	return &DisjointSet{
		parent: make(map[string]string),
		rank:   make(map[string]int),
	}
}

func (ds *DisjointSet) MakeSet(x string) {
	ds.parent[x] = x
	ds.rank[x] = 0
}

func (ds *DisjointSet) Find(x string) string {
	if ds.parent[x] != x {
		ds.parent[x] = ds.Find(ds.parent[x])
	}
	return ds.parent[x]
}

func (ds *DisjointSet) Union(x, y string) {
	rootX := ds.Find(x)
	rootY := ds.Find(y)

	if rootX != rootY {
		if ds.rank[rootX] < ds.rank[rootY] {
			ds.parent[rootX] = rootY
		} else if ds.rank[rootX] > ds.rank[rootY] {
			ds.parent[rootY] = rootX
		} else {
			ds.parent[rootY] = rootX
			ds.rank[rootX]++
		}
	}
}

func main() {
	ds := NewDisjointSet()

	// 假设 transactions 是一个包含交易数据的切片，每个交易是一个包含发送者和接收者的字符串切片
	transactions := [][]string{
		{"sender1", "receiver1"},
		{"sender2", "receiver2"},
		// 添加更多交易数据
	}

	// 初始化并查集
	for _, transaction := range transactions {
		sender := transaction[0]
		receiver := transaction[1]

		if _, exists := ds.parent[sender]; !exists {
			ds.MakeSet(sender)
		}
		if _, exists := ds.parent[receiver]; !exists {
			ds.MakeSet(receiver)
		}
	}

	// 处理交易数据并合并节点
	for _, transaction := range transactions {
		sender := transaction[0]
		receiver := transaction[1]

		ds.Union(sender, receiver)
	}

	// 计算组数
	groups := make(map[string]bool)
	for node := range ds.parent {
		root := ds.Find(node)
		groups[root] = true
	}

	numGroups := len(groups)
	fmt.Println("交易能被分成", numGroups, "组")
}
