package main

import (
	"io/ioutil"
	"net/http"
)

func procFunc(response http.ResponseWriter, req *http.Request) {
	postBody, _ := ioutil.ReadAll(req.Body)
	response.Write([]byte("query: " + req.URL.RawQuery + "\nbody: "))
	response.Write(postBody)
	response.Write([]byte("\n"))
	response.Write([]byte("backend port: 8890\n"))
}

func main() {
	http.HandleFunc("/", procFunc)
	http.ListenAndServe(":8890", nil)
}
