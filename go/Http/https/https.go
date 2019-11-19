package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"runtime/debug"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var Router *mux.Router

func SayHello(w http.ResponseWriter, r *http.Request) {
	res := map[string]string{"hello": "world"}

	b, err := json.Marshal(res)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}

func startHttpServer(address string, router *mux.Router) {
	// http server
	svr := http.Server{
		Addr:         address,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		Handler: handlers.CORS(
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS", "DELETE"}),
			handlers.AllowedOrigins([]string{"*"}))(router),
	}
	e := svr.ListenAndServe()
	if e != nil {
		fmt.Println(e.Error())
	}
}

func startHttpsServer(address string, router *mux.Router) {
	// http server
	svr := http.Server{
		Addr:         address,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		Handler: handlers.CORS(
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS", "DELETE"}),
			handlers.AllowedOrigins([]string{"*"}))(router),
	}
	e := svr.ListenAndServeTLS("server.crt", "server.key")
	if e != nil {
		fmt.Println(e.Error())
	}
}

func main() {
	// coredump stack
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()

	// http router
	Router = mux.NewRouter()
	Router.HandleFunc("/hello", SayHello).Methods("GET")

	go startHttpServer(":7040", Router)
	startHttpsServer(":7041", Router)
}
