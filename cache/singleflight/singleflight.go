package singleflight

import "sync"

type call struct { // 对key请求的抽象
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex       // 保护m
	m  map[string]*call // key到请求的抽象
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok { // 如果已经有对key的请求，复用该请求
		g.mu.Unlock()
		c.wg.Wait() // 等待执行结束
		return c.val, c.err
	}
	c := new(call) // 无请求，第一次请求
	c.wg.Add(1)
	g.m[key] = c // 暂存该请求
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done() // 运行结束，信号量-1，wait同时返回

	g.mu.Lock()
	delete(g.m, key) // 删掉暂存的请求
	g.mu.Unlock()

	return c.val, c.err
}
