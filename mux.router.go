package express

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

var _ Router = &MuxRouter{}

type MuxRouter struct {
	mux         *http.ServeMux
	inline      bool
	handler     http.Handler
	middlewares []func(http.Handler) http.Handler

	tree *RouterTree
}

func newMuxRouter() *MuxRouter {
	mux := &MuxRouter{
		mux:  http.NewServeMux(),
		tree: NewRouterTree(),
	}

	return mux
}

func (mx *MuxRouter) HFunc(pattern string, handler http.HandlerFunc) {
	parts := strings.SplitN(pattern, " ", 2)

	if len(parts) != 2 || parts[1][0] != '/' {
		panic(fmt.Sprintf("express: routing pattern must begin with method name and '/' in '%s', eg. 'GET /'", pattern))
	}

	if !mx.inline && mx.handler == nil {
		mx.updateRouteHandler()
	}

	var h http.Handler
	if mx.inline {
		mx.handler = http.HandlerFunc(mx.routeHTTP)
		h = Chain(mx.middlewares...).Handler(handler)
	} else {
		h = handler
	}

	mx.tree.Insert(parts[1], parts[0], h)
}

func (mx *MuxRouter) M(middlewares ...func(http.Handler) http.Handler) Router {
	if !mx.inline && mx.handler == nil {
		mx.updateRouteHandler()
	}

	var mws Middlewares
	if mx.inline {
		mws = make(Middlewares, len(mx.middlewares))
		copy(mws, mx.middlewares)
	}

	mws = append(mws, middlewares...)

	im := &MuxRouter{
		mux: mx.mux, inline: true, middlewares: mws, handler: mx.handler, tree: mx.tree,
	}

	return im
}

type statusCapturer struct {
	http.ResponseWriter
	status int
}

func (s *statusCapturer) WriteHeader(statusCode int) {
	s.status = statusCode
	s.ResponseWriter.WriteHeader(statusCode)
}

func (mx *MuxRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	capturer := &statusCapturer{ResponseWriter: w}
	// Create a handler that represents the final route handler
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the route information from the router tree
		routeInfo := mx.tree.Find(r.URL.Path, r.Method)

		// Check if the route is found
		if routeInfo.Path == "" {
			http.NotFound(w, r)
			return
		}

		// Check if the method is allowed for the route
		if routeInfo.Method == "" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Serve the request using the route handler
		routeInfo.Handler.ServeHTTP(w, r)
	})

	// Execute middleware chain
	handler := chain(mx.middlewares, finalHandler)

	// Process the request using the middleware chain
	handler.ServeHTTP(capturer, r)

	duration := time.Since(start)
	ms := duration.Milliseconds()

	msg := fmt.Sprintf("%v | %dms | %s | %s | %s", capturer.status, ms, strings.SplitN(r.RemoteAddr, ":", 2)[0], r.Method, r.URL.Path)

	NewLogger().Info(msg)
}

func (mx *MuxRouter) U(middlewares ...func(http.Handler) http.Handler) {
	mx.middlewares = append(mx.middlewares, middlewares...)
}

func (mx *MuxRouter) Group(fn func(r Router)) Router {
	im := mx.M()
	if fn != nil {
		fn(im)
	}
	return im
}

func (mx *MuxRouter) updateRouteHandler() {
	mx.handler = chain(mx.middlewares, http.HandlerFunc(mx.routeHTTP))
}

func (mx *MuxRouter) routeHTTP(w http.ResponseWriter, r *http.Request) {
	routePath := ""
	if routePath == "" {
		if r.URL.RawPath != "" {
			routePath = r.URL.RawPath
		} else {
			routePath = r.URL.Path
		}
		if routePath == "" {
			routePath = "/"
		}
	}

	routeInfo := mx.tree.Find(routePath, r.Method)

	if routeInfo.Path == "" {
		mx.ServeHTTP(w, r)
		return
	}

	if routeInfo.Method == "" {
		mx.ServeHTTP(w, r)
		return
	}

	routeInfo.Handler.ServeHTTP(w, r)
}
