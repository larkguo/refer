package main

import (
	"fmt"
	"log"
	"net/http"
)

// HTTPMiddleWare is a stack of Middleware Handlers that can be invoked as an http.Handler.
// HTTPMiddleWare is evaluated in the order that they are added to the stack using the Use and UseHandler methods.
type HTTPMiddleWare struct {
	middleware middleware
	handlers   []Handler
}

// New returns a new HttpMiddleWare instance with no middleware preconfigured.
func NewHTTPMiddleware(handlers ...Handler) *HTTPMiddleWare {
	return &HTTPMiddleWare{
		handlers:   handlers,
		middleware: build(handlers),
	}
}
func (n *HTTPMiddleWare) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	n.middleware.ServeHTTP(rw, r)
}

// Handler handler is an interface that objects can implement to be registered to serve as middleware
// in the HTTPMiddleWare middleware stack.
// ServeHTTP should yield to the next middleware in the chain by invoking the next http.HandlerFunc passed in.
// If the Handler writes to the ResponseWriter, the next http.HandlerFunc should not be invoked.
type Handler interface {
	ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as HTTPMiddleWare handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler object that calls f.
type HandlerFunc func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)

func (h HandlerFunc) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	h(rw, r, next)
}

type middleware struct {
	handler Handler
	nextfn  func(rw http.ResponseWriter, r *http.Request) // nextfn stores the next.ServeHTTP to reduce memory allocate
}

func newMiddleware(handler Handler, next *middleware) middleware {
	return middleware{
		handler: handler,
		nextfn:  next.ServeHTTP,
	}
}
func (m middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	m.handler.ServeHTTP(rw, r, m.nextfn)
}

func voidMiddleware() middleware {
	return newMiddleware(
		HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {}),
		&middleware{},
	)
}

func build(handlers []Handler) middleware {
	var next middleware
	switch {
	case len(handlers) == 0:
		return voidMiddleware()
	case len(handlers) > 1:
		next = build(handlers[1:])
	default:
		next = voidMiddleware()
	}
	return newMiddleware(handlers[0], &next)
}

// Wrap converts a http.Handler into a HTTPMiddleWare.Handler so it can be used as a HTTPMiddleWare
// middleware. The next http.HandlerFunc is automatically called after the Handler is executed.
func Wrap(handler http.Handler) Handler {
	return HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		handler.ServeHTTP(rw, r)
		next(rw, r)
	})
}

func main() {
	http.Handle("/hello", NewHTTPMiddleware(
		HandlerFunc(f1),
		HandlerFunc(f2),
		Wrap(http.HandlerFunc(f3)),
	))
	log.Println("listening 88")
	err := http.ListenAndServe(":88", nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
func f1(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fmt.Println("f1")
	fmt.Fprintln(w, "f1")
	next(w, r)
}
func f2(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fmt.Println("f2")
	fmt.Fprintln(w, "f2")
	next(w, r)
}
func f3(w http.ResponseWriter, r *http.Request) {
	fmt.Println("f3")
	fmt.Fprintln(w, "f3")
}
