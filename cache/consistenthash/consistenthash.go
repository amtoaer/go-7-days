package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct { // 一致性哈希实现
	hash     Hash           // 哈希函数
	replicas int            // 虚拟节点个数
	keys     []int          // 哈希环
	hashMap  map[int]string //虚拟节点与真实节点的映射
}

func New(replicas int, hash Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     hash,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE // 默认哈希函数
	}
	return m
}

func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ { // 对每个key添加replicas个虚拟节点
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))                  // 计算key的哈希
	id := sort.Search(len(m.keys), func(i int) bool { // 在数组中找到哈希值大于等于当前元素的位置
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[id%len(m.keys)]] // 使用模运算排除hash最大，id==len(m.keys)的情况
}
