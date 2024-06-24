package http

import (
	"fmt"
	"net/http"
	"strconv"
)

func SetupHttp() {

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/status", statusHandler)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	codeParams, ok := r.URL.Query()["code"]
	if ok && len(codeParams) > 0 {
		statusCode, _ := strconv.Atoi(codeParams[0])
		if statusCode >= 200 && statusCode < 600 {
			w.WriteHeader(statusCode)
		}
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	fmt.Fprintf(w, "Hello! you've requested %s\n", r.URL.Path)
}

func logRequest(r *http.Request) {
	uri := r.RequestURI
	method := r.Method
	fmt.Println("Got request!", method, uri)
}
