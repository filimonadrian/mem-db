package router

import (
	"net/http"
)

type Router struct {
	routes map[string]map[string]http.HandlerFunc
}

func NewRouter() *Router {
	router := &Router{
		routes: make(map[string]map[string]http.HandlerFunc),
	}

	return router
}

// add CORS headers to the response
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Origin, application/json")

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func (r *Router) AddRoute(method, path string, handlerFunc http.HandlerFunc) {
	if r.routes[path] == nil {
		r.routes[path] = make(map[string]http.HandlerFunc)
	}
	r.routes[path][method] = handlerFunc
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if handlers, ok := r.routes[req.URL.Path]; ok {
			if handlerFunc, methodExists := handlers[req.Method]; methodExists {
				handlerFunc(w, req)
				return
			}
		}
		http.NotFound(w, req)
	}))

	handler.ServeHTTP(w, req)
}
