package gee

import (
	"net/http"
	"strings"
)

type HandlerFunc func(c *Context)

type Engine struct {
	*RouterGroup
	groups []*RouterGroup
}

func (e *Engine) addRouter(method method, pattern string, handler HandlerFunc) {
	e.RouterGroup.router.addRouter(method, pattern, handler)
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
	var middlewares []HandlerFunc
	for _, g := range e.groups {
		if strings.HasPrefix(p, g.prefix) {
			//add middlewares
			middlewares = append(middlewares, g.getGroupMiddlewares()...)
		}
	}
	c.handlers = middlewares
	e.handle(c)
}
func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}
func New() *Engine {
	e := &Engine{}
	e.RouterGroup = newRouterGroup()
	e.RouterGroup.engine = e
	e.groups = []*RouterGroup{e.RouterGroup}
	return e
}
