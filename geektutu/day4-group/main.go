package main

import (
	"gee"
	"log"
	"net/http"
	"time"
)

// 测试功能用，尽在v2对应的Group使用
func onlyForV2() gee.HandleFunc {
	return func(c *gee.Context) {
		// start timer
		t := time.Now()
		// if a server error occurred
		//c.Fail(500, "Internet server error")	这行是毒瘤
		// Calalate resolution time
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

func main() {
	//这里用了gee的话前面就看不到后面的engine了
	r := gee.New()
	r.Use(gee.Logger()) //global middleware
	//r.GET("/index", func(c *gee.Context) {
	//	c.HTML(http.StatusOK, "<h1>Index Page</h1>")
	//})
	//v1 := r.Group("/v1")
	//{
	//	v1.GET("/", func(c *gee.Context) {
	//		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	//	})
	//
	//	v1.GET("/hello", func(c *gee.Context) {
	//		//expect /hello?name = geektutu
	//		c.String(http.StatusOK, "Hello %s, you're at %s\n", c.Query("name"), c.Path)
	//	})
	//}
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	})

	v2 := r.Group("/v2")
	v2.Use(onlyForV2())
	{
		v2.GET("/hello/:name", func(c *gee.Context) {
			//expect /hello/geektutu

			c.String(http.StatusOK, "Hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		//v2.POST("/login", func(c *gee.Context) {
		//	c.Json(http.StatusOK, gee.H{
		//		"username": c.PostForm("username"),
		//		"password": c.PostForm("password"),
		//	})
		//})
	}

	//注意：这里要换端口号的，不然会和之前的代码撞端口
	r.Run(":9996")
}

//$ curl "http://localhost:9999/hello/geektutu"
//hello geektutu, you're at /hello/geektutu
//
//$ curl "http://localhost:9999/assets/css/geektutu.css"
//{"filepath":"css/geektutu.css"}
