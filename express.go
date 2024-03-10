package express

import "net/http"

type Middlewares []func(http.Handler) http.Handler

type Router interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
	Use(middlewares ...func(http.Handler) http.Handler)
	With(middlewares ...func(http.Handler) http.Handler) Router

	Group(fn func(r Router)) Router
}

func NewRouter() *MuxRouter {
	return newMuxRouter()
}
