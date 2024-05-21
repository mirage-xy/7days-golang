package geecache

import (
	"fmt"
	"geecache/singleflight"
	"log"
	"sync"
)

// A GETTER LOADS DATA FOR A KEY
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function
// 定义函数类型GetterFunc ，实现Getter接口的Get方法，在下面
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

//函数类型实现某一个接口，称之为接口型函数，
//方便使用者在调用时既能够传入函数作为参数，也能传入实现了该接口的结构体作为参数

// A Group is a cache namespace and associated data loaded spread over
// 一个group可以认为是一个缓存的命名空间，每个group拥有一个唯一的名称name
// 比如可以创建三个group，缓存学生的成绩命名为scores，缓存学生信息的命名为info，缓存学生课程的命名为courses
type Group struct {
	name string
	//缓存未命中时获取源数据的回调（callback）
	getter Getter
	// 一开始实现的并发缓存
	mainCache cache
	// 存储将要访问的远程地址
	peers PeerPicker
	// use singleflight.Group to make sure that
	// each key is only fetched once
	loader *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup creates a new instance of group
// 实例化group，并且将group存储在全局变量groups中
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup,
// or nil if there's no such group
// 用于特定名称的Group，这里只使用只读锁RLock()，因为不涉及任何冲突变量的写操作
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get value for a key from cache
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCACHE] HIT")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	//each key is only fetched once(either locally or remotely)
	// regardless of the number of concurrent callers
	// 使用g.loader.DO包裹起来，确保并发场景下针对相同的key，load只会调用一次
	// 因为Do函数中的fn函数只会调用一次。
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			// 使用PickPeer()方法选择节点
			if peer, ok := g.peers.PickPeer(key); ok {
				// 根据节点获取对应的节点服务器
				//若非本机节点，则调用getFromPeer()获取远程节点
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCACHE] Failed to get from peer", err)
			}
		}
		// 若是本机节点或失败，则退回getLocally()
		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}
	return
}

// 使用实现了PeerGetter接口的httpGetter从访问远程节点，获取缓存值
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// RegisterPeers registers a PeerPicker for choosing remote peers
// 将实现了PeerPicker接口的HTTPPool注入到Group中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

//Get 方法实现了上述所说的流程 ⑴ 和 ⑶。
//流程 ⑴ ：从 mainCache 中查找缓存，如果存在则返回缓存值。
//流程 ⑶ ：缓存不存在，则调用 load 方法，
//load 调用 getLocally（分布式场景下会调用 getFromPeer 从其他节点获取），
//getLocally 调用用户回调函数 g.getter.Get() 获取源数据，
//并且将源数据添加到缓存 mainCache 中（通过 populateCache 方法）
