package utils

type Set map[string]struct{}

// 添加元素到集合
func (s Set) Add(item string) {
	s[item] = struct{}{}
}

// 检查集合中是否包含某个元素
func (s Set) Contains(item string) bool {
	_, found := s[item]
	return found
}

// 从集合中移除元素
func (s Set) Remove(item string) {
	delete(s, item)
}

// 获取集合中元素的数量
func (s Set) Size() int {
	return len(s)
}

// 获取集合中所有元素的切片
func (s Set) Items() []string {
	var items []string
	for item := range s {
		items = append(items, item)
	}
	return items
}
