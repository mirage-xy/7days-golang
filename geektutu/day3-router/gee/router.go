package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandleFunc
}

//root key eg , roots['GET']roots['POST']
//handlers key eg , handlers['GET-/p/:lang/doc'], handlers['POST-/p/book']

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandleFunc),
	}
}

// only one * is allowed，从路由中截取节点，只能有一个*，所有匹配到则返回
// 将长路由拆分为短的路径
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' { //若匹配到*则说明下面没有路径了，直接加入parts切片后结束
				break
			}
		}
	}
	return parts
}

// 添加路由，这里的pattern是写代码的人设定的，比如我设定的为/p/hello
// 则用户访问的时候可以输入/p/hello/xxx，这是path，也是请求路径
// 这里就要分清楚设定的路由路径pattern和用户请求路径path的区别
func (r *router) addRouter(method string, pattern string, handler HandleFunc) {
	parts := parsePattern(pattern)

	key := method + "-" + pattern
	_, ok := r.roots[method] //检查是否有一个与method对应的键，没有则创建新的node
	if !ok {                 //若不存在该key，则创建
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler //添加路由处理函数
}

// 获取路由
func (r *router) getRouter(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok { //未查询到和方法对应的路由节点，直接返回空
		return nil, nil
	}

	n := root.search(searchParts, 0)

	if n != nil {
		parts := parsePattern(n.pattern) //匹配成功的才有pattern
		for index, part := range parts {
			if part[0] == ':' { //将路由里面有：的和请求路径中的相匹配
				params[part[1:]] = searchParts[index]
			}
			//将路由里面有*的和请求路径中的相匹配，因为*后面的包括全部，所以用到Join
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break //有*的后面就不用查找了，退出即可
			}
		}
		return n, params //返回最终查找到的节点和路由与请求路径的映射map
	}

	return nil, nil //若节点n为nil，说明查找失败，返回nil

}

func (r *router) handle(c *Context) {
	n, params := r.getRouter(c.Method, c.Path)

	if n != nil {
		c.Params = params
		//上述定义中的key就是method加上pattern
		key := c.Method + "-" + n.pattern
		r.handlers[key](c) //将请求路由和处理函数绑定
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND:%s\n", c.Path)
	}
}

//router.go的变化比较小，比较重要的一点是，在调用匹配到的handler前，
//将解析出来的路由参数赋值给了c.Params。这样就能够在handler中，通过Context对象访问到具体的值了。

//我们使用 roots 来存储每种请求方式的Trie 树根节点。
//使用 handlers 存储每种请求方式的 HandlerFunc 。g
//etRoute 函数中，还解析了:和*两种匹配符的参数，返回一个 map 。
//例如/p/go/doc匹配到/p/:lang/doc，解析结果为：{lang: "go"}，
///static/css/geektutu.css匹配到/static/*filepath，解析结果为{filepath: "css/geektutu.css"}。
