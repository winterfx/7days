package gee

import (
	"net/http"
	"strings"
	"sync"
)

type router struct {
	methodsRouter map[method]*tireNode
	mux           sync.Mutex
}

func (r *router) handle(ctx *Context) {
	handler := r.searchHandler(ctx.getMethod(), ctx.getPath())
	if handler != nil {
		ctx.handlers = append(ctx.handlers, handler)
	} else {
		ctx.handlers = append(ctx.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND")
		})
	}
	ctx.Next()
}

func (r *router) addRouter(m method, pattern string, handle HandlerFunc) {
	r.mux.Lock()
	var rootN *tireNode
	if root, ok := r.methodsRouter[m]; !ok {
		//creat the root node
		rootN = &tireNode{
			childNodes: make([]*tireNode, 0),
		}
		r.methodsRouter[m] = rootN
	} else {
		rootN = root
	}
	r.mux.Unlock()
	pt := strings.Split(pattern, "/")
	newNodeList := make([]*tireNode, 0)
	for i, p := range pt {
		if len(p) == 0 {
			continue
		}
		var newNode *tireNode
		if i == len(pt)-1 {
			newNode = newtireNode(p, handle)
		} else {
			newNode = newtireNode(p, nil)
		}
		newNodeList = append(newNodeList, newNode)
	}
	if len(newNodeList) == 0 {
		rootN.handler = handle
	} else {
		rootN.insertNode(newNodeList, 0)
	}

}
func (r *router) searchHandler(m method, path string) HandlerFunc {
	r.mux.Lock()
	var rootN *tireNode
	if root, ok := r.methodsRouter[m]; !ok {
		return nil
	} else {
		rootN = root
	}
	r.mux.Unlock()
	p := strings.Split(path, "/")
	pl := make([]string, 0)
	for _, pp := range p {
		if len(pp) == 0 {
			continue
		}
		pl = append(pl, pp)
	}
	if n := rootN.searchNode(pl, 0); n != nil {
		return n.handler
	}
	return nil
}
func newRouter() *router {
	return &router{methodsRouter: make(map[method]*tireNode)}
}

///////router group combine router//////////////

type RouterGroup struct {
	prefix      string // 支持叠加
	middlewares []HandlerFunc
	engine      *Engine
	*router
}

func (r *RouterGroup) Group(prefix string) *RouterGroup {
	newGroup := &RouterGroup{
		prefix:      r.prefix + prefix,
		router:      r.router,
		middlewares: r.middlewares,
		engine:      r.engine,
	}
	r.engine.groups = append(r.engine.groups, newGroup)
	return newGroup
}

func (r *RouterGroup) Use(handler ...HandlerFunc) {
	r.middlewares = append(r.middlewares, handler...)
}
func (r *RouterGroup) addRoute(method method, comp string, handler HandlerFunc) {
	pattern := r.prefix + comp
	r.router.addRouter(method, pattern, handler)
}
func (r *RouterGroup) getGroupMiddlewares() []HandlerFunc {
	return r.middlewares
}

// GET defines the method to add GET request
func (r *RouterGroup) GET(pattern string, handler HandlerFunc) {
	r.addRoute(GET, pattern, handler)
}

// POST defines the method to add POST request
func (r *RouterGroup) POST(pattern string, handler HandlerFunc) {
	r.addRoute(POST, pattern, handler)
}
func newRouterGroup() *RouterGroup {
	return &RouterGroup{
		router:      newRouter(),
		middlewares: make([]HandlerFunc, 0),
	}
}
