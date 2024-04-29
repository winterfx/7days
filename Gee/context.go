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
	w http.ResponseWriter
	r *http.Request
}

func (c *Context) getMethod() method {
	return method(c.r.Method)
}
func (c *Context) getPath() string {
	return c.r.URL.Path
}
func (c *Context) setStatus(code int) {
	c.w.WriteHeader(code)
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
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		w: w,
		r: r,
	}
}
