package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sil-org/netsuite-mock/domain"
	"github.com/sil-org/netsuite-mock/models"
	"github.com/sil-org/netsuite-mock/storage"
)

func assertEqual[T any](t *testing.T, name string, got, want T) {
	t.Helper()
	if any(got) != any(want) {
		t.Errorf("%s: expected %v, got %v", name, want, got)
	}
}

func TestCreateCustomerHandler(t *testing.T) {
	store := storage.NewInMemoryStore()
	h := NewHandler(store)

	tests := []struct {
		name           string
		body           any
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid customer",
			body: CreateCustomerRequest{
				FirstName:  "John",
				LastName:   "Doe",
				ExternalID: "EXT123",
				Subsidiary: Related{ID: 1},
			},
			expectedStatus: http.StatusNoContent,
			expectedError:  "",
		},
		{
			name: "missing firstName",
			body: CreateCustomerRequest{
				LastName:   "Doe",
				ExternalID: "EXT456",
				Subsidiary: Related{ID: 1},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  domain.ErrorCodeMissingRequiredField,
		},
		{
			name: "missing lastName",
			body: CreateCustomerRequest{
				FirstName:  "Jane",
				ExternalID: "EXT789",
				Subsidiary: Related{ID: 1},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  domain.ErrorCodeMissingRequiredField,
		},
		{
			name: "duplicate externalId",
			body: CreateCustomerRequest{
				FirstName:  "Jane",
				LastName:   "Smith",
				ExternalID: "EXT123", // Already used
				Subsidiary: Related{ID: 1},
			},
			expectedStatus: http.StatusConflict,
			expectedError:  domain.ErrorCodeDuplicateExternalID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/services/rest/record/v1/customer", bytes.NewReader(body))
			w := httptest.NewRecorder()

			h.CreateCustomer(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusNoContent {
				location := w.Header().Get("Location")
				if location == "" {
					t.Error("expected Location header for 204 No Content")
				}
			}

			if tt.expectedError != "" {
				var response models.ErrorResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				if err != nil {
					t.Errorf("error decoding response body: %v", err)
				}
				if response.Error.Code != tt.expectedError {
					t.Errorf("expected error %s, got %s", tt.expectedError, response.Error.Code)
				}
			}
		})
	}
}

func TestGetCustomerHandler(t *testing.T) {
	store := storage.NewInMemoryStore()
	h := NewHandler(store)

	// Create a customer first with all supported fields
	customer := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "EXT123",
		Subsidiary: 1,
		IsInactive: false,
		AddressBook: &models.AddressBook{
			Items: []models.AddressBookItem{
				{
					AddressBookAddress: &models.AddressBookAddress{
						Country: "US",
					},
				},
			},
		},
	}
	_, _ = store.CreateCustomer(context.Background(), customer)

	// Get the customer
	req := httptest.NewRequest(http.MethodGet, "/services/rest/record/v1/customer/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	h.GetCustomer(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response models.Customer
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	assertEqual(t, "ID", response.ID, int64(1))
	assertEqual(t, "FirstName", response.FirstName, "John")
	assertEqual(t, "LastName", response.LastName, "Doe")
	assertEqual(t, "ExternalID", response.ExternalID, "EXT123")
	assertEqual(t, "Subsidiary", response.Subsidiary, int64(1))
	assertEqual(t, "IsInactive", response.IsInactive, false)

	if response.AddressBook == nil || len(response.AddressBook.Items) == 0 {
		t.Fatal("expected AddressBook items, but got none")
	}
	assertEqual(t, "Country", response.AddressBook.Items[0].AddressBookAddress.Country, "US")
}

func TestCreateInvoiceHandler(t *testing.T) {
	store := storage.NewInMemoryStore()
	h := NewHandler(store)

	// Create a customer first
	customer := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		Subsidiary: 1,
	}
	created, _ := store.CreateCustomer(context.Background(), customer)

	tests := []struct {
		name           string
		body           any
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid invoice",
			body: CreateInvoiceRequest{
				Entity:     Related{ID: created.ID},
				Date:       "2024-01-31T00:00:00Z",
				Subsidiary: Related{ID: 1},
				Currency:   Related{ID: 1},
			},
			expectedStatus: http.StatusNoContent,
			expectedError:  "",
		},
		{
			name: "non-existent customer",
			body: CreateInvoiceRequest{
				Entity:     Related{ID: 999},
				Date:       "2024-01-31T00:00:00Z",
				Subsidiary: Related{ID: 1},
				Currency:   Related{ID: 1},
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  domain.ErrorCodeRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/services/rest/record/v1/invoice", bytes.NewReader(body))
			w := httptest.NewRecorder()

			h.CreateInvoice(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusNoContent {
				location := w.Header().Get("Location")
				if location == "" {
					t.Error("expected Location header for 204 No Content")
				}
			}

			if tt.expectedError != "" {
				var response models.ErrorResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				if err != nil {
					t.Errorf("error decoding response body: %v", err)
				}
				if response.Error.Code != tt.expectedError {
					t.Errorf("expected error %s, got %s", tt.expectedError, response.Error.Code)
				}
			}
		})
	}
}

func TestCreateJournalEntryHandler(t *testing.T) {
	store := storage.NewInMemoryStore()
	h := NewHandler(store)

	tests := []struct {
		name           string
		body           any
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid journal entry",
			body: CreateJournalEntryRequest{
				Memo:       "Test Entry",
				Subsidiary: Related{ID: 1},
				Line: &JournalLinesWrapper{
					Items: []JournalLine{
						{Debit: 100},
						{Credit: 100},
					},
				},
			},
			expectedStatus: http.StatusNoContent,
			expectedError:  "",
		},
		{
			name: "debit credit mismatch",
			body: CreateJournalEntryRequest{
				Memo:       "Bad Entry",
				Subsidiary: Related{ID: 1},
				Line: &JournalLinesWrapper{
					Items: []JournalLine{
						{Debit: 100},
						{Credit: 50},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  domain.ErrorCodeDebitCreditMismatch,
		},
		{
			name: "missing memo",
			body: CreateJournalEntryRequest{
				Memo:       "",
				Subsidiary: Related{ID: 1},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  domain.ErrorCodeMissingRequiredField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/services/rest/record/v1/journalEntry", bytes.NewReader(body))
			w := httptest.NewRecorder()

			h.CreateJournalEntry(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusNoContent {
				location := w.Header().Get("Location")
				if location == "" {
					t.Error("expected Location header for 204 No Content")
				}
			}

			if tt.expectedError != "" {
				var response models.ErrorResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				if err != nil {
					t.Errorf("error decoding response body: %v", err)
				}
				if response.Error.Code != tt.expectedError {
					t.Errorf("expected error %s, got %s", tt.expectedError, response.Error.Code)
				}
			}
		})
	}
}

func TestQueryHandler(t *testing.T) {
	store := storage.NewInMemoryStore()
	h := NewHandler(store)

	// Create some test data
	customer := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "JOHN_DOE",
		Subsidiary: 1,
		IsInactive: false,
	}
	_, err := store.CreateCustomer(context.Background(), customer)
	if err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}

	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedError  string
		shouldHaveData bool
		expectedData   map[string]any
	}{
		{
			name: "valid customer query",
			query: "SELECT id,firstName,lastName,externalId,subsidiary,isInactive " +
				"FROM customer WHERE externalId = 'JOHN_DOE'",
			expectedStatus: http.StatusOK,
			expectedError:  "",
			shouldHaveData: true,
			expectedData: map[string]any{
				"id":         "1",
				"firstname":  "John",
				"lastname":   "Doe",
				"externalid": "JOHN_DOE",
				"subsidiary": float64(1),
				"isinactive": "F",
			},
		},
		{
			name:           "invalid query syntax",
			query:          "INVALID QUERY",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_QUERY_SYNTAX",
			shouldHaveData: false,
		},
		{
			name:           "unsupported where operator",
			query:          "SELECT id FROM customer WHERE id > 0",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_QUERY_SYNTAX",
			shouldHaveData: false,
		},
		{
			name:           "unrecognized filter field",
			query:          "SELECT id FROM customer WHERE unknownField = 'x'",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_QUERY_FILTER",
			shouldHaveData: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(QueryRequest{Q: tt.query})
			req := httptest.NewRequest(http.MethodPost, "/services/rest/query/v1/suiteql", bytes.NewReader(body))
			w := httptest.NewRecorder()

			h.QueryHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.shouldHaveData {
				var resp models.QueryResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if resp.Count == 0 {
					t.Error("expected data in response, but count is 0")
				}

				if len(resp.Items) == 0 {
					t.Fatal("expected items in response, but items list is empty")
				}

				item := resp.Items[0]
				for k, v := range tt.expectedData {
					got, ok := item[k]
					if !ok {
						t.Errorf("missing expected field %s in response", k)
						continue
					}
					assertEqual(t, "field "+k, got, v)
				}
			} else if tt.expectedError != "" {
				var response models.ErrorResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				if err != nil {
					t.Fatalf("error decoding response body: %v", err)
				}
				if response.Error == nil {
					t.Fatalf("expected error response body")
				}
				if response.Error.Code != tt.expectedError {
					t.Errorf("expected error %s, got %s", tt.expectedError, response.Error.Code)
				}
			}
		})
	}
}

func TestGetInvoiceItemHandler(t *testing.T) {
	store := storage.NewInMemoryStore()
	h := NewHandler(store)

	// Create a customer first
	customer := &models.Customer{
		FirstName:  "Jane",
		LastName:   "Smith",
		Subsidiary: 1,
	}
	created, err := store.CreateCustomer(context.Background(), customer)
	if err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}

	// Create an invoice with two line items
	invoice := models.NewInvoice()
	invoice.Entity = models.Entity{ID: created.ID}
	invoice.Subsidiary = models.Subsidiary{ID: 1}
	invoice.Currency = models.Currency{ID: 1}
	invoice.Items = []models.InvoiceItem{
		{Item: 101, Amount: 50.00, Description: "First item"},
		{Item: 102, Amount: -10.00, Description: "Credit item"},
	}
	createdInvoice, err := store.CreateInvoice(context.Background(), invoice)
	if err != nil {
		t.Fatalf("failed to create invoice: %v", err)
	}

	invoiceIDStr := fmt.Sprintf("%d", createdInvoice.ID)

	tests := []struct {
		name           string
		invoiceID      string
		lineNum        string
		expectedStatus int
		expectedError  string
		checkItem      func(t *testing.T, item models.InvoiceItem)
	}{
		{
			name:           "get first line item",
			invoiceID:      invoiceIDStr,
			lineNum:        "1",
			expectedStatus: http.StatusOK,
			checkItem: func(t *testing.T, item models.InvoiceItem) {
				assertEqual(t, "Item", item.Item, int64(101))
				assertEqual(t, "Amount", item.Amount, 50.00)
				assertEqual(t, "Description", item.Description, "First item")
			},
		},
		{
			name:           "get second line item (negative amount)",
			invoiceID:      invoiceIDStr,
			lineNum:        "2",
			expectedStatus: http.StatusOK,
			checkItem: func(t *testing.T, item models.InvoiceItem) {
				assertEqual(t, "Item", item.Item, int64(102))
				assertEqual(t, "Amount", item.Amount, -10.00)
				assertEqual(t, "Description", item.Description, "Credit item")
			},
		},
		{
			name:           "line number out of range",
			invoiceID:      invoiceIDStr,
			lineNum:        "3",
			expectedStatus: http.StatusNotFound,
			expectedError:  domain.ErrorCodeRecordNotFound,
		},
		{
			name:           "line number zero",
			invoiceID:      invoiceIDStr,
			lineNum:        "0",
			expectedStatus: http.StatusNotFound,
			expectedError:  domain.ErrorCodeRecordNotFound,
		},
		{
			name:           "non-numeric line number",
			invoiceID:      invoiceIDStr,
			lineNum:        "abc",
			expectedStatus: http.StatusBadRequest,
			expectedError:  domain.ErrorCodeInvalidRequest,
		},
		{
			name:           "invoice not found",
			invoiceID:      "99999",
			lineNum:        "1",
			expectedStatus: http.StatusNotFound,
			expectedError:  domain.ErrorCodeRecordNotFound,
		},
		{
			name:           "non-numeric invoice ID",
			invoiceID:      "abc",
			lineNum:        "1",
			expectedStatus: http.StatusBadRequest,
			expectedError:  domain.ErrorCodeInvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/services/rest/record/v1/invoice/%s/item/%s", tt.invoiceID, tt.lineNum)
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.SetPathValue("id", tt.invoiceID)
			req.SetPathValue("lineNum", tt.lineNum)
			w := httptest.NewRecorder()

			h.GetInvoiceItem(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkItem != nil {
				var item models.InvoiceItem
				if err := json.NewDecoder(w.Body).Decode(&item); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				tt.checkItem(t, item)
			}

			if tt.expectedError != "" {
				var response models.ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("error decoding response body: %v", err)
				}
				if response.Error == nil {
					t.Fatalf("expected error response body")
				}
				if response.Error.Code != tt.expectedError {
					t.Errorf("expected error %s, got %s", tt.expectedError, response.Error.Code)
				}
			}
		})
	}
}
