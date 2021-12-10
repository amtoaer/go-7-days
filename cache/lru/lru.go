package lru

import "container/list"

// Cache 最基础的lru缓存实现
type Cache struct {
	maxBytes  int64                    // 最大字节
	nbytes    int64                    // 当前字节
	ll        *list.List               // 双向链表
	cache     map[string]*list.Element // key->链表元素映射
	OnEvicted func(string, Value)      // 删除时触发的函数
}

type entry struct {
	key   string // 键
	Value Value  // 值
}

type Value interface {
	Len() int // 实现Len方法即可作为缓存的值
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		OnEvicted: onEvicted,
	}
}

func (c Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.Value, true
	}
	return nil, false
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() // 最不常访问
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.Value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.Value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { // 元素已经存在
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.Value.Len())
		kv.Value = value
	} else {
		ele := c.ll.PushFront(&entry{key: key, Value: value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
