package gee

import (
	"net/http"
)

type handler func(*Context)

type Gee interface {
	Get(string, handler)
	Post(string, handler)
	Run(string)
}

type gee struct {
	router *router
}

var _ http.Handler = &gee{}

func New() Gee {
	return &gee{router: &router{route: make(map[string]handler)}}
}

func (g *gee) addRoute(method, pattern string, fun handler) {
	g.router.addRoute(method, pattern, fun)
}

func (g *gee) Get(pattern string, fun handler) {
	g.addRoute("GET", pattern, fun)
}

func (g *gee) Post(pattern string, fun handler) {
	g.addRoute("POST", pattern, fun)
}

func (g *gee) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	context := &Context{writer: writer, request: request}
	g.router.handle(context)
}

func (g *gee) Run(addr string) {
	http.ListenAndServe(addr, g)
}
