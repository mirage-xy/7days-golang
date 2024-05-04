package main

import (
	"gee"
	"net/http"
)

func main() {
	//这里用了gee的话前面就看不到后面的engine了
	r := gee.New()

	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "<h1>hello gee</h1>")
	})

	r.GET("/hello", func(c *gee.Context) {
		c.String(http.StatusOK, "hello %s,you are at %s\n", c.Query("name"), c.Path)
	})

	//localhost:9998/login?username=wzy&password=123
	r.POST("/login", func(c *gee.Context) {
		c.Json(http.StatusOK, gee.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})
	//注意：这里要换端口号的，不然会和之前的代码撞端口
	r.Run(":9998")
}
