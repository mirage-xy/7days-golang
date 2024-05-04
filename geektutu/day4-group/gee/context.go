package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//注意：write用于处理响应体，writeHeader用于处理响应头
//对于web请求来说，无非是根据请求*http.request，构造响应http.responseWriter。
//在HandlerFunc中，我们希望能够访问到解析的参数，因此，需要对Context对象增加一个属性和一个方法
//来提供对路由参数的访问，讲解析后的参数存储到Params中，通过c.Params("lang")的方式获取到对应的值

type H map[string]interface{}

type Context struct {
	//初始对象
	Writer http.ResponseWriter
	Req    *http.Request
	//请求信息
	Path   string            //请求路径
	Method string            //请求方法
	Params map[string]string //路由与请求路径的映射map，即解析后的参数
	//响应信息
	StatusCode int //响应状态码
	//middleware
	handlers []HandleFunc
	index    int //记录当前执行到第几个中间件
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	//此处状态码是响应信息，现在不能确定，就先不定义，此处用Context来接受请求信息
	return &Context{
		Writer: w,
		Req:    r,
		Path:   r.URL.Path,
		Method: r.Method,
		index:  -1,
	}
}

// 这是递归的过程，在中间件中调用next方法时，控制权交给下一个中间件，直到调用到最后一个中间件
// 然后在从后往前，调用每个中间件在Next方法之后定义的部分。
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.Json(code, H{"message": err})
}

// 开始定义Context有关的方法
// 访问PostForm参数的方法
func (c *Context) PostForm(key string) string {
	//FormValue方法返回请求中根据key得到的第一个value
	//如果不存在key，返回nil
	return c.Req.FormValue(key)
}

// 查找的方法
// c.Req.URL 是请求的 URL 对象，Query()方法返回一个URL的查询参数的 url.Values 映射，
// 然后 Get(key) 从这个映射中获取与键（key）对应的值。
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// 设置状态码，用于向客户端发送HTTP响应的状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// 设置请求头
func (c *Context) SetHeader(key string, value string) {
	//ResponseWriter 接口通常有一个 Header 方法，
	//它返回一个 http.Header 类型的值，该值是一个映射，用于存储 HTTP 响应头的键值对。
	//在 http.Header 映射上调用 Set 方法来设置指定的键（key）和值（value）。
	//如果键已经存在，它的值将被新的值替换；如果键不存在，将添加一个新的键值对。
	c.Writer.Header().Set(key, value)
}

// 快速构造String/Data/JSON/HTML响应的方法。
// 方便讲不同类型的数据作为HTTP响应发送回客户端，同时设置适当的状态码和Content-Type头
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain") //表示响应体是纯文本
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...))) //将字符串写入响应体
}

func (c *Context) Json(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json") //表示响应体是json格式
	c.Status(code)
	//创建一个JSON编码器，该编码器可以将go中的数据结构编码为JSON字节流
	//并且该编码器以http.ResponseWriter为输入
	//后面可以调用该编码器将数据结构转化为json数据，并且写入HTTP响应中。
	encoder := json.NewEncoder(c.Writer)
	//这里开始编码并写入响应
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

//设计context的必要性
//1、对Web服务来说，无非是根据请求*http.Request，构造响应http.ResponseWriter。
//但是这两个对象提供的接口粒度太细，比如我们要构造一个完整的响应，
//需要考虑消息头(Header)和消息体(Body)，而 Header 包含了状态码(StatusCode)，
//消息类型(ContentType)等几乎每次请求都需要设置的信息。
//因此，如果不进行有效的封装，那么框架的用户将需要写大量重复，繁杂的代码，而且容易出错。
//针对常用场景，能够高效地构造出 HTTP 响应是一个好的框架必须考虑的点。
//2、针对使用场景，封装*http.Request和http.ResponseWriter的方法，简化相关接口的调用，
//只是设计 Context 的原因之一。对于框架来说，还需要支撑额外的功能。
//例如，将来解析动态路由/hello/:name，参数:name的值放在哪呢？
//再比如，框架需要支持中间件，那中间件产生的信息放在哪呢？
//Context 随着每一个请求的出现而产生，请求的结束而销毁，
//和当前请求强相关的信息都应由 Context 承载。
//因此，设计 Context 结构，扩展性和复杂性留在了内部，而对外简化了接口。
//路由的处理函数，以及将要实现的中间件，参数都统一使用 Context 实例，
//Context 就像一次会话的百宝箱，可以找到任何东西。
