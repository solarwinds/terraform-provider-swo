package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func ConvertObject[T any](from any) (*T, error) {
	b, err := json.Marshal(&from)
	if err != nil {
		return nil, err
	}

	var result T
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}

	return &result, err
}

// A debugging function which dumps the HTTP response to stdout.
func DumpResponse(resp *http.Response) {
	fmt.Printf("response status: %s\n", resp.Status)
	dump, err := httputil.DumpResponse(resp, true)

	if err != nil {
		log.Printf("error dumping response: %s", err)
		return
	}

	log.Printf("response body: %s\n\n", string(dump))
}

// A debugging function which dumps the HTTP request to stdout.
func DumpRequest(req *http.Request) {
	if req.Body == nil {
		return
	}

	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Printf("error dumping request: %s", err)
		return
	}

	log.Printf("request body: %s\n\n", string(dump))
}

func ArrayElementExists(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
