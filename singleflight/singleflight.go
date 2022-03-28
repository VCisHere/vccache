package singleflight

import "sync"

// 表示正在进行中或已经结束的请求
type call struct {
	wg sync.WaitGroup
	val interface{}
	err error
}

// 管理不同key的请求
type Group struct {
	mu sync.Mutex
	m map[string]*call
}

// 针对相同的key，无论Do被调用多少次，fn都只调用一次，等待fn调用结束了，返回
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()				// 如果请求正在进行中，则等待
		return c.val, c.err		// 请求结束，返回结果
	}
	c := new(call)
	c.wg.Add(1)			// 发起请求前加锁
	g.m[key] = c				// 表明key已经有请求正在处理

	c.val, c.err = fn()			// 调用fn，发起请求
	c.wg.Done()					// 请求结束

	delete(g.m, key)			// key的请求处理完毕

	return c.val, c.err
}