package cache

import (
	"fmt"
	"log"
	"sync"

	"github.com/amtoaer/go-7-days/cache/lru"
)

// cache 对lru的进一步封装，为操作加锁
type cache struct {
	mu         sync.Mutex // 互斥锁
	lru        *lru.Cache // lru缓存
	cacheBytes int64      // 缓存大小
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error) // 抽象一个缓存未命中时获取数据的方法

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string     // 名称
	getter    Getter     // 未命中时获取数据的方法
	mainCache cache      // 本地缓存
	peers     PeerPicker // 远程节点的选择器
}

var (
	groups = make(map[string]*Group) // 名称->group的映射
	mu     sync.RWMutex              // 用于修改上述映射的读写锁
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" { // key为空报错
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok { // 首先尝试使用本地缓存
		log.Println("[Cache] hit")
		return v, nil
	}
	return g.load(key) // 尝试其它方法
}

func (g *Group) RegisterPeers(peers PeerPicker) { // 注册节点选择器
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil { // 如果节点选择器存在，尝试远程缓存
		if peer, ok := g.peers.PickPeer(key); ok { // 通过一致性哈希获取key对应的节点数据获取器
			if value, err = g.getFromPeer(peer, key); err == nil { // 尝试从该节点获取缓存的数据
				return value, err
			}
			log.Println("[Cache]Failed to get from peer", err) // 获取失败
		}
	}
	return g.getLocally(key) // 本机及远程缓存均不存在，本地拉取新缓存
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (value ByteView, err error) {
	bytes, err := peer.Get(g.name, key) //使用节点数据获取器调取远程缓存
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bytes}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key) // 使用本地getter函数获取key对应的value值
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{copySlice(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) { // 在本地缓存内添加数据
	g.mainCache.add(key, value)
}
