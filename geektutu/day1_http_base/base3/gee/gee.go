package gee

import (
	"fmt"
	"net/http"
)

type handleFunc func(w http.ResponseWriter, r *http.Request)

// 有个engine以后相当于所有的HTTP请求都被我们拦截
// 拥有统一的控制入口，即要通过engine对路由进行控制
// Engine实现ServeHTTP接口，相当于实现了Handle接口
type Engine struct {
	//路由，目前猜测：key指的是请求方法和静态路由地址构成，handleFunc指的是处理函数
	//请求方法：get,pos
	router map[string]handleFunc
}

// 新建一个Engine结构体对象
func NewEngine() *Engine {
	return &Engine{router: make(map[string]handleFunc)}
}

// 添加路由
func (engine *Engine) addRoute(method string, pattern string, handler handleFunc) {
	//这里就构造了一个路由，key由请求方法和路由地址构成，value由处理函数构成
	key := method + "-" + pattern
	engine.router[key] = handler
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
	key := r.Method + "-" + r.URL.Path
	//这个语句前面声明handle和ok,分号后面判断ok是否为true
	if handler, ok := engine.router[key]; ok {
		//找到处理函数，然后把处理函数的参数w和r传过去
		handler(w, r)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", r.URL)
	}
}
