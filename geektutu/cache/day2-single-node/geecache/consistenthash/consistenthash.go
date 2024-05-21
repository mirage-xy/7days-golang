package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// HASH maps bytes to uint32
type Hash func(data []byte) uint32

// Map constains all hashed keys
type Map struct {
	hash     Hash
	replicas int            // 虚拟节点倍数
	keys     []int          //哈希环
	hashMap  map[int]string //虚拟节点与真实节点的映射表
}

func New(replicas int, hash Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     hash,
		hashMap:  make(map[int]string),
	}
	//如果没有自定义哈希算法，则迷默认为下列算法
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

// 添加服务器节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			//这里不用string()是因为string得到的是ascal码值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key //将虚拟节点与真实节点映射
		}
	}
	sort.Ints(m.keys)
}

// 选择节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
