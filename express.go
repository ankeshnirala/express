package express

import "net/http"

type Middlewares []func(http.Handler) http.Handler

type Router interface {
	HFunc(pattern string, handler http.HandlerFunc)
	U(middlewares ...func(http.Handler) http.Handler)
	M(middlewares ...func(http.Handler) http.Handler) Router

	Group(fn func(r Router)) Router
}

func NewRouter() *MuxRouter {
	return newMuxRouter()
}
