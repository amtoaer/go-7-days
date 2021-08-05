package gee

import "fmt"

type router struct {
	route map[string]handler
}

func (r *router) addRoute(method, pattern string, handle handler) {
	key := fmt.Sprintf("%s-%s", method, pattern)
	r.route[key] = handle
}

func (r *router) handle(c *Context) {
	key := fmt.Sprintf("%s-%s", c.request.Method, c.request.URL.Path)
	if fun, ok := r.route[key]; ok {
		fun(c)
	} else {
		c.notFound()
	}
}
