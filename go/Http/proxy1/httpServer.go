package main

import (
	"fmt"
	"net/http"
	"os"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("<- ", r.RemoteAddr, r.RequestURI)
	fmt.Fprintln(w, "receive request from:", r.RemoteAddr, r.RequestURI)
}

// http server
func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: ./server listenIP:listenPort")
		return
	}
	listen_addr := os.Args[1]
	http.HandleFunc("/", IndexHandler)
	fmt.Println("Listen ", listen_addr)

	err := http.ListenAndServe(listen_addr, nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}