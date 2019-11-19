/*
curl --data "hello http proxy"  http://127.0.0.1:8888/?param=test
*/

package main

import (
	"io/ioutil"
	"net/http"
)

func proxyFunc(response http.ResponseWriter, req *http.Request) {
	client := &http.Client{}
	path := req.URL.Path
	query := req.URL.RawQuery
	url := "http://127.0.0.1:8890"
	url += path
	if len(query) > 0 {
		url += "?" + query
	}

	proxyReq, err := http.NewRequest("POST", url, req.Body)
	if err != nil {
		response.Write([]byte("http proxy request fail\n"))
		return
	}

	proxyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	proxyReq.Header.Set("Cookie", "name=cookie")

	resp, err := client.Do(proxyReq)
	defer resp.Body.Close()
	out, _ := ioutil.ReadAll(resp.Body)
	response.Write(out)
}

func main() {
	http.HandleFunc("/", proxyFunc)
	go http.ListenAndServe(":8888", nil)
	http.ListenAndServeTLS(":8889", "server.crt", "server.key", nil)
}
