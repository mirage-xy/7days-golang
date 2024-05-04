package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

type HandleFunc func(c *Context)

// 整个框架的所有资源都是由Engine统一协调的，你们就可以通过Engine间接的访问各种接口
// Engine实现ServeHTTP接口，相当于实现了Handle接口
type Engine struct {
	//engine实例和路由相关联，即拦截HTTP请求，所以其中的属性是路由，用来接收HTTP请求
	router *router
	*RouterGroup
	groups        []*RouterGroup     //存储所有分组
	htmlTemplates *template.Template //for html render,将所有的模板加载进内存
	funcMap       template.FuncMap   //for hmtl render，自定义的模版渲染函数
}

type RouterGroup struct {
	prefix      string       //前缀
	middlewares []HandleFunc //support middleware
	engine      *Engine      //all group share an Engine instance
}

// 新建一个Engine结构体对象
func New() *Engine {
	//这里开始创建新的engine
	engine := &Engine{
		router: newRouter(),
	}
	engine.RouterGroup = &RouterGroup{
		engine: engine,
	}
	engine.groups = []*RouterGroup{
		engine.RouterGroup,
	}

	//返回创建的engine
	return engine
}

// group is defined to creat a new RouterGroup
// remember all groups share the same Egine instance
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix, //子路由前缀加上现有的路由前缀，才是完整的路由路径
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// default use Logger() & Recovery middlewares
func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery()) //global middleware
	return engine
}

// 加载渲染函数
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

// 加载模板
func (engine *Engine) LoadHTMLGlob(pattern string) {
	//使用html/template包来加载并解析HTML模板文件
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandleFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		//check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// serve static files
// 这个方法是暴露给用户的，用户可以将磁盘上某个文件夹root映射到路由relativePath
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	//register GET handlers,这样用户就不调用get而是调用static方法了
	group.GET(urlPattern, handler)
}

// 此处是通过engine添加路由的代码
// 添加路由
func (engine *Engine) addRoute(method string, pattern string, handler HandleFunc) {
	//这里就构造了一个路由，将与路由相关的都转义到router中，这里只负责调用方法
	engine.router.addRouter(method, pattern, handler)
}

// 添加get请求
func (engine *Engine) GET(pattern string, handler HandleFunc) {
	engine.addRoute("GET", pattern, handler)
}

// 添加post请求
func (engine *Engine) POST(pattern string, handler HandleFunc) {
	engine.addRoute("POST", pattern, handler)
}

// 此后是通过group组添加路由的代码
// 添加路由
func (group *RouterGroup) addRoute(method string, comp string, handler HandleFunc) {
	//这里就构造了一个路由，将与路由相关的都转义到router中，这里只负责调用方法
	pattern := group.prefix + comp
	log.Printf("router %4s - %s", method, pattern)
	group.engine.router.addRouter(method, pattern, handler)
}

// 添加get请求
func (group *RouterGroup) GET(pattern string, handler HandleFunc) {
	group.addRoute("GET", pattern, handler)
}

// 添加post请求
// 这里不能写group.engine.addRouter，因为这样就不是使用组添加了
func (group *RouterGroup) POST(pattern string, handler HandleFunc) {
	group.addRoute("POST", pattern, handler)
}

// 开启HTTP服务。就是那个监听函数
func (engine *Engine) Run(addr string) error {
	//这里engine要先实现ServeHTTP方法，不然没有实现Handle接口，传不过去
	return http.ListenAndServe(addr, engine)
}

// Use被定义用来向组内添加中间件
func (group *RouterGroup) Use(middlewares ...HandleFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// engine实现ServeHTTP方法，这里的作用是解析请求的路径，根据路径去查找路由表，即查找map
func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//但现在查找路由这一部分让独立出来的router去做
	var middlewares []HandleFunc
	for _, group := range engine.groups {
		// strings.HasPrefix()函数用于检查一个字符串是否以制定的前缀开始
		// 如果URL.Path是以group.prefix开头，表示这个请求应该应用该路由组的中间件，
		//如果不是以该前缀开头，则不使用改组中间件
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, r)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}
