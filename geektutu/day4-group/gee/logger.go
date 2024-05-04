package gee

import (
	"log"
	"time"
)

//中间件是应用在RouterGroup上的，应用在最顶层的 Group，相当于作用于全局，所有的请求都会被中间件处理

func Logger() HandleFunc {
	return func(c *Context) {
		//start timer
		t := time.Now()
		//process request
		c.Next()
		//Calculate resolution time
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
