package main

import (
	"fmt"
	"gee"
	"html/template"
	"net/http"
	"time"
)

type student struct {
	Name string
	Age  int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	//这里用了gee的话前面就看不到后面的engine了
	r := gee.Default()

	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	stu1 := &student{Name: "Geektutu", Age: 20}
	stu2 := &student{Name: "Jack", Age: 22}
	r.GET("/", func(c *gee.Context) {
		//c.HTML(http.StatusOK, "css.tmpl", nil)
		c.String(http.StatusOK, "hello geektutu\n")
	})

	//index out of range for testing recovery
	r.GET("/panic", func(c *gee.Context) {
		names := []string{"geektutu"}
		c.String(http.StatusOK, names[100])
	})

	r.GET("/student", func(c *gee.Context) {
		c.HTML(http.StatusOK, "arr.tmpl", gee.H{
			"title":  "gee",
			"stuArr": [2]*student{stu1, stu2},
		})
	})

	r.GET("/date", func(c *gee.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", gee.H{
			"title": "gee",
			"now":   time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
		})
	})

	r.Run(":9995")
}

//$ curl "http://localhost:9999/hello/geektutu"
//hello geektutu, you're at /hello/geektutu
//
//$ curl "http://localhost:9999/assets/css/geektutu.css"
//{"filepath":"css/geektutu.css"}
