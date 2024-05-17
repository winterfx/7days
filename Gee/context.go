package gee

import (
	"encoding/json"
	"net/http"
)

type method string

const (
	GET  method = "GET"
	POST method = "POST"
)

type Context struct {
	w        http.ResponseWriter
	r        *http.Request
	code     int
	handlers []HandlerFunc
	index    int
}

func (c *Context) getMethod() method {
	return method(c.r.Method)
}
func (c *Context) getPath() string {
	return c.r.URL.Path
}
func (c *Context) GetStatusCode() int {
	return c.code
}
func (c *Context) setStatus(code int) {
	c.w.WriteHeader(code)
	c.code = code
}
func (c *Context) setHeader(key, value string) {
	c.w.Header().Set(key, value)
}
func (c *Context) JSON(code int, obj interface{}) {
	c.setHeader("Content-Type", "application/json")
	c.setStatus(code)
	encoder := json.NewEncoder(c.w)
	if err := encoder.Encode(obj); err != nil {
		panic(err.Error())
	}

}
func (c *Context) String(code int, s string) {
	c.setHeader("Content-Type", "text/plain")
	c.setStatus(code)
	c.w.Write([]byte(s))
}
func (c *Context) Fail(code int, s string) {
	c.String(code, s)
}
func (c *Context) Next() {
	c.index++
	l := len(c.handlers)
	//https://geektutu.com/post/gee-day5.html#:~:text=Login%20with%20GitHub-,hu%2Dxiaokangcommentedover%204%20years%20ago,-func%20(c
	//不是所有的handler都会调用 Next()。
	//手工调用 Next()，一般用于在请求前后各实现一些行为。如果中间件只作用于请求前，可以省略调用Next()，算是一种兼容性比较好的写法吧。
	for ; c.index < l; c.index++ {
		c.handlers[c.index](c)
	}
}
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		w:        w,
		r:        r,
		handlers: make([]HandlerFunc, 0),
		index:    -1,
	}
}
