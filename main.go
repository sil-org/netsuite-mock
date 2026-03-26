package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sil-org/netsuite-mock/handlers"
	"github.com/sil-org/netsuite-mock/storage"
)

func requestLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Parse command-line flags
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	// Initialize storage
	store := storage.NewInMemoryStore()
	defer store.Close(context.Background())

	// Initialize handlers
	h := handlers.NewHandler(store)

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Query endpoint
	mux.HandleFunc("POST /services/rest/query/v1/suiteql", h.QueryHandler)

	// Customer endpoints
	mux.HandleFunc("POST /services/rest/record/v1/customer", h.CreateCustomer)
	mux.HandleFunc("GET /services/rest/record/v1/customer/{id}", h.GetCustomer)
	mux.HandleFunc("PATCH /services/rest/record/v1/customer/{id}", h.UpdateCustomer)

	// Employee endpoints
	mux.HandleFunc("POST /services/rest/record/v1/employee", h.CreateEmployee)
	mux.HandleFunc("GET /services/rest/record/v1/employee/{id}", h.GetEmployee)

	// Account endpoints
	mux.HandleFunc("POST /services/rest/record/v1/account", h.CreateAccount)
	mux.HandleFunc("GET /services/rest/record/v1/account/{id}", h.GetAccount)

	// Invoice endpoints
	mux.HandleFunc("POST /services/rest/record/v1/invoice", h.CreateInvoice)
	mux.HandleFunc("GET /services/rest/record/v1/invoice/{id}", h.GetInvoice)
	mux.HandleFunc("GET /services/rest/record/v1/invoice/{id}/item/{lineNum}", h.GetInvoiceItem)

	// Customer Payment endpoints
	mux.HandleFunc("POST /services/rest/record/v1/customerPayment", h.CreatePayment)
	mux.HandleFunc("GET /services/rest/record/v1/customerPayment/{id}", h.GetPayment)

	// Journal Entry endpoints
	mux.HandleFunc("POST /services/rest/record/v1/journalEntry", h.CreateJournalEntry)
	mux.HandleFunc("GET /services/rest/record/v1/journalEntry/{id}", h.GetJournalEntry)
	mux.HandleFunc("GET /services/rest/record/v1/journalEntry/{id}/line/{lineNum}", h.GetJournalEntryLine)

	mux.HandleFunc("/", handlers.InvalidURL)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      requestLogMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting NetSuite mock server on port %s\n", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
