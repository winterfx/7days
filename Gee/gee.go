package gee

import (
	"net/http"
)

type HandlerFunc func(c *Context)

type Engine struct {
	router *router
}

func (e *Engine) addRouter(method method, pattern string, handler HandlerFunc) {
	e.router.addRouter(method, pattern, handler)
}
func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRouter(GET, pattern, handler)
}
func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.addRouter(POST, pattern, handler)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r)
	p := c.r.URL.Path
	f := e.router.searchHandler(c.getMethod(), p)
	if f == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("NOT FOUND"))
	} else {
		f(c)
	}

}
func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}
func New() *Engine {
	return &Engine{
		router: newRouter(),
	}
}
