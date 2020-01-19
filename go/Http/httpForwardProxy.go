package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

type Proxy struct {
}

func main() {
	// coredump stack
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()

	// proxy start
	server := &http.Server{
		Addr:    ":8888",
		Handler: &Proxy{},
	}
	fmt.Println(server.ListenAndServe())
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		handleTunneling(w, r)
	} else {
		handleHttp(w, r)
	}
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	// 1.build proxyRequest
	proxyReq := new(http.Request)
	proxyReq = r
	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		prior, ok := proxyReq.Header["X-Forwarded-For"]
		if ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		proxyReq.Header.Set("X-Forwarded-For", clientIP)
	}

	// 2.transfer request
	resp, err := http.DefaultTransport.RoundTrip(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// 3.transfer response
	for key, value := range resp.Header {
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body) // stream copy
}

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source) // stream copy
}
