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

type Proxy struct { // inherit  Handler.ServeHTTP
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
		Handler: &Proxy{}, //  start goroutine for every request
	}
	fmt.Println(server.ListenAndServe())
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1.build channel
	doneCh := make(chan struct{})
	timeoutCh := make(chan struct{})
	defer func() {
		close(doneCh)
		close(timeoutCh)
	}()

	// 2.timing
	go func() {
		time.Sleep(32 * time.Second) // RoundTrip timeout
		timeoutCh <- struct{}{}
	}()

	// 3.proxy request
	if r.Method == http.MethodConnect {
		go proxyHttps(doneCh, w, r)
	} else {
		go proxyHttp(doneCh, w, r)
	}

	// 4.wait
	ctx := r.Context()
	select {
	case <-doneCh: // proxy done
		return
	case <-timeoutCh: // timeout
		return
	case <-ctx.Done(): // client cancel
		err := ctx.Err()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func proxyHttp(doneCh chan<- struct{}, w http.ResponseWriter, r *http.Request) {
	defer func() {
		doneCh <- struct{}{}
	}()

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
	resp, err := http.DefaultTransport.RoundTrip(proxyReq) // proxyReq.Context().Done()
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

func proxyHttps(doneCh chan<- struct{}, w http.ResponseWriter, r *http.Request) {
	defer func() {
		doneCh <- struct{}{}
	}()

	dest_conn, err := net.Dial("tcp", r.Host) // net.DialTimeout
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
