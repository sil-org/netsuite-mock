package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sil-org/netsuite-mock/models"
	"github.com/sil-org/netsuite-mock/storage"
)

const (
	HeaderContentType = "Content-Type"
	ContentTypeJSON   = "application/json"
)

// Handler contains the HTTP handlers for the API
type Handler struct {
	store storage.Store
}

// NewHandler creates a new handler with the provided store
func NewHandler(store storage.Store) *Handler {
	return &Handler{
		store: store,
	}
}

// respondError sends an error response to the client
func respondError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(statusCode)

	log.Println("Error", statusCode, code, message)

	response := models.NewErrorResponse(code, message, "CLIENT_ERROR")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("ERROR (respondError): %s", err)
	}
}

// respondNoContent sends a 204 No Content response with a Location header
func respondNoContent(w http.ResponseWriter, r *http.Request, id string) {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	location := fmt.Sprintf("%s://%s%s/%s", scheme, r.Host, r.URL.RequestURI(), id)
	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusNoContent)
}

// respondSuccess sends a successful response to the client
func respondSuccess(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("ERROR (respondSuccess): %s", err)
	}
}

// respondQuery sends a query response to the client
func respondQuery(w http.ResponseWriter, statusCode int, items []map[string]any) {
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(statusCode)

	response := models.NewQueryResponse()
	response.Items = items
	response.Count = len(items)
	response.TotalResults = len(items)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("ERROR (respondQuery): %s", err)
	}
}
