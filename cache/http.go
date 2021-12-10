package cache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/amtoaer/go-7-days/cache/consistenthash"
)

const (
	defaultBasePath = "/_cache/" // path前缀
	defaultReplicas = 50         // 默认一个节点对应的虚拟节点个数
)

type HTTPPool struct {
	self        string                 // 该节点地址
	basePath    string                 // path前缀
	peers       *consistenthash.Map    // 一致性哈希map实现
	httpGetters map[string]*HttpGetter //
	mu          sync.Mutex             // 互斥锁
}

var _ PeerPicker = (*HTTPPool)(nil)

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) { // 需要有同样的path前缀
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	args := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2) // 从公共path前缀后截断，通过/分割，得到groupName和key
	if len(args) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName, key := args[0], args[1]
	group := GetGroup(groupName) // 找到name对应的Group
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key) // 使用Group的本地缓存获取缓存值
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice()) // 将结果写入body
}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil) // 初始化默认虚拟节点数的一致性哈希
	p.peers.Add(peers...)                              // 将节点列表添加
	p.httpGetters = make(map[string]*HttpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &HttpGetter{baseURL: peer + p.basePath} // 初始化httpGetter表
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) { // 通过key得到对应节点的getter，可直接调取其get方法获取缓存
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peers %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}
