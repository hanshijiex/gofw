package gofw

import "net/http"

type Gofw struct {
	Router *Router
	Server *http.Server
}

func NewGoFw() *Gofw {
	Router := NewRouter()
	Gofw := &Gofw{Router: Router, Server: &http.Server{}}

	return Gofw
}

func (gofw *Gofw) AddRoute(method, pattern string, handlerFunc HandlerFunc) {
	gofw.Router.Add(method, pattern, handlerFunc)
}

func (gofw *Gofw) Listen(port string) error {
	gofw.Server.Handler = gofw.Router
	gofw.Server.Addr = port

	return gofw.Server.ListenAndServe()
}