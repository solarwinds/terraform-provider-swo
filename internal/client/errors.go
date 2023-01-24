package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorResponse represents the response body returned when the API reports an error.
type ErrorResponse struct {
	Errors   interface{} `json:"errors"`
	Status   string      `json:"status"`
	Response *http.Response
}

// Error makes ErrorResponse satisfy the error interface and can be used to serialize
// error responses back to the httpClient.
func (e *ErrorResponse) Error() string {
	errorData, _ := json.Marshal(e.Errors)
	return fmt.Sprintf("%s - %s", e.Status, string(errorData))
}
