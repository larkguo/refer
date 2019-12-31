package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func proxyLogin(w http.ResponseWriter, r *http.Request) {
	trueServer := "http://127.0.0.1:8003"
	url, err := url.Parse(trueServer)
	if err != nil {
		log.Println(err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w, r)
}
func proxyDefault(w http.ResponseWriter, r *http.Request) {
	trueServer := "http://127.0.0.1:8890"
	url, err := url.Parse(trueServer)
	if err != nil {
		log.Println(err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w, r)

}

// https 0.0.0.0:8889/login -> http 127.0.0.1 8003
// https 0.0.0.0:8889/ -> http 127.0.0.1 8890
func main() {
	http.HandleFunc("/login", proxyLogin)
	http.HandleFunc("/", proxyDefault)
	http.ListenAndServeTLS(":8889", "server.crt", "server.key", nil)
}