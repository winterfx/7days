package gee

import (
	"strings"
	"sync"
)

type router struct {
	methodsRouter map[method]*routerNode
	mux           sync.Mutex
}

type routerNode struct {
	partPattern string
	handler     HandlerFunc
	childNodes  []*routerNode
	isWild      bool //':' or '*' is wild
}

func (r *routerNode) matchChild(newNode *routerNode) *routerNode {
	for _, n := range r.childNodes {
		if n.partPattern == newNode.partPattern || n.isWild {
			return n
		}
	}
	return nil
}
func (r *routerNode) matchChildren(part string) []*routerNode {
	nodes := make([]*routerNode, 0)
	for _, child := range r.childNodes {
		if child.partPattern == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}
func (r *routerNode) insertNode(nodeList []*routerNode, height int) {
	if height == len(nodeList) {
		return
	}
	newNode := nodeList[height]
	child := r.matchChild(newNode)
	if child == nil {
		r.childNodes = append(r.childNodes, newNode)
	}
	child.insertNode(nodeList, height+1)
}
func (r *routerNode) searchNode(path []string, height int) *routerNode {
	if len(path) == height || strings.HasPrefix(r.partPattern, "*") {
		return r
	}
	p := path[height]
	children := r.matchChildren(p)

	for _, child := range children {
		result := child.searchNode(path, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}

func newRouterNode(partPattern string, handler HandlerFunc) *routerNode {
	r := &routerNode{
		partPattern: partPattern,
		childNodes:  make([]*routerNode, 0),
		handler:     handler,
	}
	if strings.HasPrefix(partPattern, ":") || strings.HasPrefix(partPattern, "*") {
		r.isWild = true
	}
	return r
}
func (r *router) addRouter(m method, pattern string, handle HandlerFunc) {
	//r.mux.Lock()
	var rootN *routerNode
	if root, ok := r.methodsRouter[m]; !ok {
		//creat the root node
		rootN = &routerNode{
			childNodes: make([]*routerNode, 0),
		}
		r.methodsRouter[m] = rootN
	} else {
		rootN = root
	}
	//r.mux.Unlock()
	pt := strings.Split(pattern, "/")
	newNodeList := make([]*routerNode, 0)
	for i, p := range pt {
		if len(p) == 0 {
			continue
		}
		var newNode *routerNode
		if i == len(pt)-1 {
			newNode = newRouterNode(p, handle)
		} else {
			newNode = newRouterNode(p, nil)
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
	var rootN *routerNode
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
	return &router{methodsRouter: make(map[method]*routerNode)}
}
