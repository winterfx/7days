package gee

import "strings"

type tireNode struct {
	partPattern string
	handler     HandlerFunc
	childNodes  []*tireNode
	isWild      bool //':' or '*' is wild
}

func (r *tireNode) insertNode(nodeList []*tireNode, height int) {
	if height == len(nodeList) {
		return
	}
	newNode := nodeList[height]
	child := r.matchChild(newNode)
	if child == nil {
		r.childNodes = append(r.childNodes, newNode)
		child = newNode
	}
	child.insertNode(nodeList, height+1)
}
func (r *tireNode) searchNode(path []string, height int) *tireNode {
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
func (r *tireNode) matchChild(newNode *tireNode) *tireNode {
	for _, n := range r.childNodes {
		if n.partPattern == newNode.partPattern || n.isWild {
			return n
		}
	}
	return nil
}

func (r *tireNode) matchChildren(part string) []*tireNode {
	nodes := make([]*tireNode, 0)
	for _, child := range r.childNodes {
		if child.partPattern == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}
func newtireNode(partPattern string, handler HandlerFunc) *tireNode {
	r := &tireNode{
		partPattern: partPattern,
		childNodes:  make([]*tireNode, 0),
		handler:     handler,
	}
	if strings.HasPrefix(partPattern, ":") || strings.HasPrefix(partPattern, "*") {
		r.isWild = true
	}
	return r
}
