package main

import (
	"fmt"
	"gee"
	"net/http"
)

func main() {
	r := gee.NewEngine()

	//添加路由
	r.GET("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "URL.PATH=%Q\n", r.URL.Path)
	})

	r.GET("/hello", func(w http.ResponseWriter, r *http.Request) {
		//因为header其实是一个map，这里key拿到header的键，v拿到值，然后输出
		for k, v := range r.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	})

	r.Run(":9999")

}
