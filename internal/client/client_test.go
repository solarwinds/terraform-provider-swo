package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/google/go-cmp/cmp"
)

const (
	baseURLPath = "/graphql"
)

// Sets up a test HTTP server along with an swo.Client that is configured to make requests
// to the test server. Tests register handlers on mux which provide mock responses for the
// API being tested.
func setup() (ctx context.Context, client *Client, mux *http.ServeMux, serverURL string, teardown func()) {
	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	apiHandler := http.NewServeMux()
	apiHandler.Handle(baseURLPath+"/", http.StripPrefix(baseURLPath, mux))
	apiHandler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "Client.BaseURL path prefix is not preserved in the request URL.", http.StatusInternalServerError)
	})

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(apiHandler)

	// client is the SWO client being tested and is configured to use the test server.
	url, _ := url.Parse(server.URL + baseURLPath + "/")
	client = NewClient("123456", BaseUrlOption(url.String()))

	return context.Background(), client, mux, server.URL, server.Close
}

func httpErrorResponse(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "You are a teapot.", http.StatusTeapot)
}

func sendGraphQLResponse(t *testing.T, w io.Writer, response any) bool {
	err := json.NewEncoder(w).Encode(graphql.Response{Data: response})
	if err != nil {
		t.Errorf("Swo.SendGraphQLResponse error: %v", err)
		return false
	}

	return true
}

func getGraphQLInput[T any](r *http.Request) (*T, error) {
	request := new(graphql.Request)
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		return nil, err
	}

	varsBytes, err := json.Marshal(request.Variables)
	if err != nil {
		return nil, err
	}

	var requestInput T
	err = json.Unmarshal(varsBytes, &requestInput)
	if err != nil {
		return nil, err
	}

	return &requestInput, err
}

func testObjects(t *testing.T, obj1 any, obj2 any) bool {
	if !cmp.Equal(obj1, obj2) {
		t.Log(cmp.Diff(obj1, obj2))
		return false
	}

	return true
}

// Test whether the marshaling of v produces JSON that corresponds
// to the want string.
func testJSONMarshal(t *testing.T, v interface{}, want string) {
	t.Helper()
	// Unmarshal the wanted JSON, to verify its correctness, and marshal it back
	// to sort the keys.
	u := reflect.New(reflect.TypeOf(v)).Interface()
	if err := json.Unmarshal([]byte(want), &u); err != nil {
		t.Errorf("Unable to unmarshal JSON for %v: %v", want, err)
	}
	w, err := json.Marshal(u)
	if err != nil {
		t.Errorf("Unable to marshal JSON for %#v", u)
	}

	// Marshal the target value.
	j, err := json.Marshal(v)
	if err != nil {
		t.Errorf("Unable to marshal JSON for %#v", v)
	}

	if string(w) != string(j) {
		t.Errorf("json.Marshal(%q) returned %s, want %s", v, j, w)
	}
}
