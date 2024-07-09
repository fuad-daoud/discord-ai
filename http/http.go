package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

func Setup() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/status", statusHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		slog.Error("Could not serve on 8080")
		panic(err)
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	codeParams, ok := r.URL.Query()["cli"]
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
