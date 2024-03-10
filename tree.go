package express

import (
	"net/http"
	"strings"
)

type RouteInfo struct {
	Path    string
	Method  string
	Handler http.Handler
}

type node struct {
	children map[string]*node
	handlers map[string]http.Handler
}

type RouterTree struct {
	root *node
}

func NewRouterTree() *RouterTree {
	return &RouterTree{root: &node{children: make(map[string]*node), handlers: make(map[string]http.Handler)}}
}

func (rt *RouterTree) Insert(path string, method string, handler http.Handler) {
	currentNode := rt.root
	pathSegments := strings.Split(path, "/")

	for _, segment := range pathSegments {
		if currentNode.children[segment] == nil {
			currentNode.children[segment] = &node{children: make(map[string]*node), handlers: make(map[string]http.Handler)}
		}
		currentNode = currentNode.children[segment]
	}

	// Store the handler for the specified method
	currentNode.handlers[method] = handler
}

func (rt *RouterTree) Find(path string, method string) *RouteInfo {
	currentNode := rt.root
	pathSegments := strings.Split(path, "/")

	for _, segment := range pathSegments {
		if currentNode.children[segment] == nil {
			// Path segment not found
			return &RouteInfo{Path: "", Method: method, Handler: nil}
		}
		currentNode = currentNode.children[segment]
	}

	// Retrieve the handler for the specified method
	handler, ok := currentNode.handlers[method]
	if !ok {
		// Method not found
		return &RouteInfo{Path: path, Method: "", Handler: nil}
	}

	// Return the path, method, and handler
	return &RouteInfo{Path: path, Method: method, Handler: handler}
}
