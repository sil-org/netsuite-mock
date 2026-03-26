package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse is the NetSuite error response format.
type ErrorResponse struct {
	Type         string         `json:"type"`
	Title        string         `json:"title"`
	Status       int            `json:"status"`
	ErrorDetails []ErrorDetails `json:"o:errorDetails"`
}

// ErrorDetails is for detailed reporting of error information in an error response.
type ErrorDetails struct {
	Detail    string `json:"detail"`
	ErrorUrl  string `json:"o:errorUrl"`
	ErrorCode string `json:"o:errorCode"`
}

// InvalidURL is the catch-all handler for unsupported URLs
func InvalidURL(w http.ResponseWriter, r *http.Request) {
	log.Println("Invalid URL: ", r.RequestURI)
	response := ErrorResponse{
		Type:   "https://www.rfc-editor.org/rfc/rfc9110.html#section-15.5.1",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		ErrorDetails: []ErrorDetails{{
			Detail: "Invalid request URL. The request URL must be constructed in the following way: " +
				"<service>/v1/<recordname>/<recordid>.",
			ErrorUrl:  r.RequestURI,
			ErrorCode: "INVALID_URL",
		}},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(response)
}
