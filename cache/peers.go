package cache

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type PeerPicker interface { // 通过key找到对应节点
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface { // 节点的缓存Getter
	Get(group string, key string) ([]byte, error)
}

type HttpGetter struct {
	baseURL string // 对于远程通信，需要有baseURL
}

var _ PeerGetter = (*HttpGetter)(nil) // 类型检查，查看是否实现接口

func (h *HttpGetter) Get(group string, key string) ([]byte, error) {
	url := fmt.Sprintf("%s%s/%s", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %v", resp.Status)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}
