package lru

import "container/list"

// Cache is a LRU cache. It is not safe for concurrent access
type Cache struct {
	maxBytes int64 //允许使用的最大内存
	nbytes   int64 //当前已使用的内存
	ll       *list.List
	cache    map[string]*list.Element //值是双向链表中对应节点的指针
	//optional and executed when an entry is purged
	//为了通用性，允许值是实现了value接口的任意类型，该接口只包含一个方法Len()返回值所占用的内存大小
	OnEvicted func(key string, value Value) //某条记录被移除时的回调函数，可以为nil
}

// 双向链表节点的数据类型，在链表中仍保存每个值对应的key的好处在于，淘汰队首节点时，需要用key从字典中删除对应映射
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 查找:主要有两个步骤，第一步是从字典中找到对应的双向链表的节点
// 第二步， 将该节点移动到队尾
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		//将链表中的节点ele移动到队尾，这里约定front为队尾
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// 删除：实际上是缓存淘汰，溢出最近最少访问的节点
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		//从cache中删除节点映射关系
		delete(c.cache, kv.key)
		//更新当前所用内存
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		//如果回调函数不为nil，则调用回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// add adds a value to the cache
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		//如果存在则代表访问了，所以添加到队尾，并且更新内容
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		//如果不存在则新加节点，并且在map这里面添加key和节点的映射关系
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	//更新c.nbytes，如果超过了设定的最大值c.maxBytes，则移除最少访问的节点
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {

	return c.ll.Len()
}
