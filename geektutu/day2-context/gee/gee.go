package gee

import (
	"net/http"
)

type handleFunc func(c *Context)

// Engine实现ServeHTTP接口，相当于实现了Handle接口
type Engine struct {
	//engine实例和路由相关联，即拦截HTTP请求，所以其中的属性是路由，用来接收HTTP请求
	router *router
}

// 新建一个Engine结构体对象
func New() *Engine {
	return &Engine{router: newRouter()}
}

// 添加路由
func (engine *Engine) addRoute(method string, pattern string, handler handleFunc) {
	//这里就构造了一个路由，将与路由相关的都转义到router中，这里只负责调用方法
	engine.router.addRouter(method, pattern, handler)
}

// 添加get请求
func (engine *Engine) GET(pattern string, handler handleFunc) {
	engine.addRoute("GET", pattern, handler)
}

// 添加post请求
func (engine *Engine) POST(pattern string, handler handleFunc) {
	engine.addRoute("POST", pattern, handler)
}

// 开启HTTP服务。就是那个监听函数
func (engine *Engine) Run(addr string) error {
	//这里engine要先实现ServeHTTP方法，不然没有实现Handle接口，传不过去
	return http.ListenAndServe(addr, engine)
}

// engine实现ServeHTTP方法，这里的作用是解析请求的路径，根据路径去查找路由表，即查找map
func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//但现在查找路由这一部分让独立出来的router去做
	c := newContext(w, r)
	engine.router.handle(c)
}
