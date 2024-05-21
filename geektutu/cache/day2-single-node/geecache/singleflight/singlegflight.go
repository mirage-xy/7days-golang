package singleflight

import "sync"

// call代表正在进行中，或已经结束的请求，使用sync.waitGroup锁避免重入
type call struct {
	// 用于同步等待fn函数的执行完成，当一个新的call被创建时，c.wg.Add(1)被调用，表示有一个新的等待项
	// 当fn函数执行完毕后，c.wg.Done()被调用，标识等待项已经完成
	// 其他正在等待这个请求完成的groutine通过c.wg.Wait()可以继续执行。
	// wg确保在fn函数执行期间，没有其他的groutine会尝试获取这个请求结果
	// 这不是锁，而是对一组线程的等待，即等待与特点call关联的fn函数执行完成
	wg  sync.WaitGroup
	val interface{}
	err error
}

// group是singleflight的主数据结果，管理不同key的请求（call）
type Group struct {
	// 互斥锁，用于保护共享资源的并发访问
	mu sync.Mutex // protects m
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 在并发环境下，多个和线程尝试读取或修改m时，如果不加锁，就可能导致数据不一致或意外行为
	g.mu.Lock() //g.mu.Lock()是保护Group的成员变量m不被并发读写而加上的锁
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok { //检查请求call是否在处理中或者已完成
		g.mu.Unlock()
		c.wg.Wait()         // 如果请求正在进行中，则等待
		return c.val, c.err // 请求结束，返回结果
	}
	c := new(call)
	c.wg.Add(1)  //发起请求前加锁
	g.m[key] = c // 添加到g.m，表明key已经有对应的请求在处理
	g.mu.Unlock()

	c.val, c.err = fn() //调用fn，发起请求
	c.wg.Done()         //请求结束

	g.mu.Lock()
	delete(g.m, key) //更新g.m
	g.mu.Unlock()

	return c.val, c.err
}
