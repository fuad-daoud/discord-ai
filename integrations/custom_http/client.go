package custom_http

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

type Client interface {
	DoJson(req *http.Request, result any)
	Do(req *http.Request) []byte
	MakeRequest(method string, path string, data *strings.Reader) *http.Request
	GetRequest(path string) *http.Request
	PostRequest(path string, data *strings.Reader) *http.Request
	PostEmptyRequest(path string) *http.Request
}

type DefaultClient struct {
	Client  *http.Client
	BaseURL string
	Headers map[string]string
}

func (dc *DefaultClient) DoJson(req *http.Request, result any) {
	body := dc.Do(req)
	if err := json.Unmarshal(body, result); err != nil {
		log.Fatalf("Can not unmarshal JSON\n%s", string(body))
	}
	//log.Printf("url %s, json: %s", req.URL.Path, string(body))
}

func (dc *DefaultClient) Do(req *http.Request) []byte {
	resp, err := dc.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode >= 400 {

		log.Fatalf("status code is more than 400 status %s url: %s\nbody: %s", resp.Status, req.URL.String(), string(body))
	}
	return body
}

func (dc *DefaultClient) MakeRequest(method string, path string, data *strings.Reader) *http.Request {
	req, err := http.NewRequest(method, dc.BaseURL+path, data)
	if err != nil {
		log.Fatal(err)
	}
	dc.setHeaders(req)
	return req
}
func (dc *DefaultClient) GetRequest(path string) *http.Request {
	req, err := http.NewRequest("GET", dc.BaseURL+path, strings.NewReader(""))
	if err != nil {
		log.Fatal(err)
	}
	dc.setHeaders(req)
	return req
}
func (dc *DefaultClient) PostRequest(path string, data *strings.Reader) *http.Request {
	req, err := http.NewRequest("POST", dc.BaseURL+path, data)
	if err != nil {
		log.Fatal(err)
	}
	dc.setHeaders(req)
	return req
}
func (dc *DefaultClient) PostEmptyRequest(path string) *http.Request {
	req, err := http.NewRequest("POST", dc.BaseURL+path, strings.NewReader(""))
	if err != nil {
		log.Fatal(err)
	}
	dc.setHeaders(req)
	return req
}
func (dc *DefaultClient) MakeRequestNoHeaders(method string, path string, data *strings.Reader) *http.Request {
	req, err := http.NewRequest(method, dc.BaseURL+path, data)
	if err != nil {
		log.Fatal(err)
	}
	return req
}

func (dc *DefaultClient) setHeaders(req *http.Request) {
	for k, v := range dc.Headers {
		req.Header.Set(k, v)
	}
}
