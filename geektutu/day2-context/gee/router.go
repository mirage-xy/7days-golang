package gee

import (
	"log"
	"net/http"
)

// 将路由相关的方法和结果提取出来，放到新的文件中
// 方便下次对router的功能进行增强。例如提供动态路由支持
// handle方法做了一个细微调整，即handler的参数，变成了context
type router struct {
	handlers map[string]handleFunc
}

func newRouter() *router {
	return &router{handlers: make(map[string]handleFunc)}
}

// pattern表示路由的模式或者路径，因为常见的路径有多种方式，不同的摸式
func (r *router) addRouter(method string, pattern string, handler handleFunc) {
	log.Printf("router %4s-%s", method, pattern)
	key := method + "-" + pattern
	r.handlers[key] = handler
}

func (r *router) handle(c *Context) {
	key := c.Method + "-" + c.Path
	if handler, ok := r.handlers[key]; ok {
		handler(c)
	} else {
		c.String(http.StatusFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
