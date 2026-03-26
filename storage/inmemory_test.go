package storage

import (
	"context"
	"testing"

	"github.com/sil-org/netsuite-mock/domain"
	"github.com/sil-org/netsuite-mock/models"
)

func TestCreateCustomer(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	customer := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "EXT123",
		Subsidiary: 1,
	}

	created, err := store.CreateCustomer(ctx, customer)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created.ID != 1 {
		t.Errorf("expected ID 1, got %d", created.ID)
	}

	if created.ExternalID != "EXT123" {
		t.Errorf("expected externalId EXT123, got %s", created.ExternalID)
	}
}

func TestCreateCustomerDuplicateExternalId(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	customer1 := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "EXT123",
		Subsidiary: 1,
	}

	customer2 := &models.Customer{
		FirstName:  "Jane",
		LastName:   "Smith",
		ExternalID: "EXT123",
		Subsidiary: 1,
	}

	_, err := store.CreateCustomer(ctx, customer1)
	if err != nil {
		t.Fatalf("expected no error for first customer, got %v", err)
	}

	_, err = store.CreateCustomer(ctx, customer2)
	if err == nil {
		t.Fatal("expected error for duplicate externalId, got nil")
	}

	if err.Error() != domain.ErrorCodeDuplicateExternalID {
		t.Errorf("expected DUPLICATE_EXTERNAL_ID error, got %v", err)
	}
}

func TestCreateCustomerEmptyExternalId(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Two customers with no externalId should be allowed
	customer1 := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "",
		Subsidiary: 1,
	}

	customer2 := &models.Customer{
		FirstName:  "Jane",
		LastName:   "Smith",
		ExternalID: "",
		Subsidiary: 1,
	}

	_, err := store.CreateCustomer(ctx, customer1)
	if err != nil {
		t.Fatalf("expected no error for first customer, got %v", err)
	}

	_, err = store.CreateCustomer(ctx, customer2)
	if err != nil {
		t.Fatalf("expected no error for second customer with empty externalId, got %v", err)
	}
}

func TestGetCustomer(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	customer := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "EXT123",
		Subsidiary: 1,
	}

	created, _ := store.CreateCustomer(ctx, customer)
	retrieved, err := store.GetCustomer(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.FirstName != "John" {
		t.Errorf("expected firstName John, got %s", retrieved.FirstName)
	}
}

func TestGetCustomerNotFound(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	_, err := store.GetCustomer(ctx, 999)
	if err == nil {
		t.Fatal("expected error for non-existent customer, got nil")
	}

	if err.Error() != domain.ErrorCodeRecordNotFound {
		t.Errorf("expected RECORD_NOT_FOUND error, got %v", err)
	}
}

func TestGetCustomerByExternalId(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	customer := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "EXT123",
		Subsidiary: 1,
	}

	_, _ = store.CreateCustomer(ctx, customer)
	retrieved, err := store.GetCustomerByExternalID(ctx, "EXT123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.FirstName != "John" {
		t.Errorf("expected firstName John, got %s", retrieved.FirstName)
	}
}

func TestExternalIDExists(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	customer := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "EXT123",
		Subsidiary: 1,
	}

	_, err := store.CreateCustomer(ctx, customer)
	if err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}

	// Should find existing externalId
	exists, err := store.ExternalIDExists(ctx, "customer", "EXT123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !exists {
		t.Error("expected externalId to exist")
	}

	// Should not find non-existent externalId
	exists, err = store.ExternalIDExists(ctx, "customer", "NOTFOUND")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if exists {
		t.Error("expected externalId to not exist")
	}

	// Empty externalId should always return false
	exists, err = store.ExternalIDExists(ctx, "customer", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if exists {
		t.Error("expected empty externalId to not exist")
	}
}

func TestCreateAccount(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	account := &models.Account{
		AccountNumber: "10701-AAB",
		Class:         "AAB",
	}

	created, err := store.CreateAccount(ctx, account)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created.ID != 1 {
		t.Errorf("expected ID 1, got %d", created.ID)
	}
}

func TestCreateInvoiceRequiresCustomer(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	invoice := &models.Invoice{
		Entity:     models.Entity{ID: 999}, // Non-existent customer
		TranDate:   "2024-01-01",
		Subsidiary: models.Subsidiary{ID: 1},
	}

	_, err := store.CreateInvoice(ctx, invoice)
	if err == nil {
		t.Fatal("expected error for non-existent customer, got nil")
	}

	if err.Error() != domain.ErrorCodeRecordNotFound {
		t.Errorf("expected RECORD_NOT_FOUND error, got %v", err)
	}
}

func TestCreateInvoiceSuccess(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Create a customer first
	customer := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "EXT123",
		Subsidiary: 1,
	}
	createdCustomer, _ := store.CreateCustomer(ctx, customer)

	invoice := &models.Invoice{
		Entity:     models.Entity{ID: createdCustomer.ID}, // Non-existent customer
		TranDate:   "2024-01-01",
		Subsidiary: models.Subsidiary{ID: 1},
	}

	created, err := store.CreateInvoice(ctx, invoice)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created.ID != 1 {
		t.Errorf("expected ID 1, got %d", created.ID)
	}

	if created.TranID != "INV-1" {
		t.Errorf("expected tranId INV-1, got %s", created.TranID)
	}
}

func TestCreateJournalEntryValidation(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	tests := []struct {
		name      string
		entry     *models.JournalEntry
		wantError string
	}{
		{
			name: "less than 2 lines",
			entry: &models.JournalEntry{
				Memo:       "Test",
				Subsidiary: models.Subsidiary{ID: 1},
				Lines: []models.JournalLine{
					{Debit: 100},
				},
			},
			wantError: domain.ErrorCodeInvalidRequest,
		},
		{
			name: "debit credit mismatch",
			entry: &models.JournalEntry{
				Memo:       "Test",
				Subsidiary: models.Subsidiary{ID: 1},
				Lines: []models.JournalLine{
					{Debit: 100},
					{Credit: 50},
				},
			},
			wantError: domain.ErrorCodeDebitCreditMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.CreateJournalEntry(ctx, tt.entry)
			if err == nil {
				t.Fatalf("expected error %s, got nil", tt.wantError)
			}
			if err.Error() != tt.wantError {
				t.Errorf("expected error %s, got %v", tt.wantError, err)
			}
		})
	}
}

func TestCreateJournalEntrySuccess(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	entry := &models.JournalEntry{
		Memo:       "Test Entry",
		Subsidiary: models.Subsidiary{ID: 1},
		Lines: []models.JournalLine{
			{Debit: 100},
			{Credit: 100},
		},
	}

	created, err := store.CreateJournalEntry(ctx, entry)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created.ID != 1 {
		t.Errorf("expected ID 1, got %d", created.ID)
	}

	if created.TranID != "JE-1" {
		t.Errorf("expected tranId JE-1, got %s", created.TranID)
	}
}

func TestCreateCustomerPaymentRequiresRecords(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	payment := &models.CustomerPayment{
		Customer:   models.Customer{ID: 999}, // Non-existent
		Account:    models.Account{ID: 999},
		Payment:    100,
		Subsidiary: models.Subsidiary{ID: 1},
	}

	_, err := store.CreateCustomerPayment(ctx, payment)
	if err == nil {
		t.Fatal("expected error for non-existent customer, got nil")
	}
}

func TestCreateCustomerPaymentSuccess(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Create customer and account
	customer := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		Subsidiary: 1,
	}
	createdCustomer, _ := store.CreateCustomer(ctx, customer)

	account := &models.Account{
		AccountNumber: "10701-AAB",
		Class:         "AAB",
	}
	createdAccount, _ := store.CreateAccount(ctx, account)

	payment := &models.CustomerPayment{
		Customer:   *createdCustomer,
		Account:    *createdAccount,
		Payment:    -100, // Can be negative
		Subsidiary: models.Subsidiary{ID: 1},
	}

	created, err := store.CreateCustomerPayment(ctx, payment)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created.ID != 1 {
		t.Errorf("expected ID 1, got %d", created.ID)
	}

	if created.Payment != -100 {
		t.Errorf("expected payment -100, got %f", created.Payment)
	}
}

func TestUpdateCustomer(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	customer := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "EXT123",
		Subsidiary: 1,
	}

	created, _ := store.CreateCustomer(ctx, customer)

	// Update to inactive
	created.IsInactive = true
	updated, err := store.UpdateCustomer(ctx, created)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !updated.IsInactive {
		t.Error("expected customer to be inactive")
	}
}

func TestUpdateCustomerExternalIdChange(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	customer := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "EXT123",
		Subsidiary: 1,
	}

	created, _ := store.CreateCustomer(ctx, customer)

	// Update external ID - create a new customer object with the same ID but new externalId
	updated := &models.Customer{
		ID:         created.ID,
		FirstName:  created.FirstName,
		LastName:   created.LastName,
		ExternalID: "EXT456",
		Subsidiary: created.Subsidiary,
		IsInactive: created.IsInactive,
	}
	updated, err := store.UpdateCustomer(ctx, updated)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.ExternalID != "EXT456" {
		t.Errorf("expected externalId EXT456, got %s", updated.ExternalID)
	}

	// Old external ID should not be found
	_, err = store.GetCustomerByExternalID(ctx, "EXT123")
	if err == nil {
		t.Fatal("expected error for old externalId, got nil")
	}
	if err.Error() != domain.ErrorCodeRecordNotFound {
		t.Errorf("expected RECORD_NOT_FOUND error, got %v", err)
	}

	// New external ID should be found
	retrieved, err := store.GetCustomerByExternalID(ctx, "EXT456")
	if err != nil {
		t.Fatalf("expected no error for new externalId, got %v", err)
	}

	if retrieved.FirstName != "John" {
		t.Errorf("expected firstName John, got %s", retrieved.FirstName)
	}
}

func TestUpdateCustomerDuplicateExternalIdDoesNotDeleteOld(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Create two customers with different externalIds
	customer1 := &models.Customer{
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "EXT123",
		Subsidiary: 1,
	}
	created1, _ := store.CreateCustomer(ctx, customer1)

	customer2 := &models.Customer{
		FirstName:  "Jane",
		LastName:   "Smith",
		ExternalID: "EXT456",
		Subsidiary: 1,
	}
	created2, _ := store.CreateCustomer(ctx, customer2)

	// Try to update customer1's externalId to customer2's externalId (should fail)
	customer1Copy := &models.Customer{
		ID:         created1.ID,
		FirstName:  created1.FirstName,
		LastName:   created1.LastName,
		ExternalID: "EXT456", // This is already used by customer2
		Subsidiary: created1.Subsidiary,
		IsInactive: created1.IsInactive,
	}

	_, err := store.UpdateCustomer(ctx, customer1Copy)
	if err == nil {
		t.Fatal("expected error when updating to duplicate externalId, got nil")
	}
	if err.Error() != domain.ErrorCodeDuplicateExternalID {
		t.Errorf("expected DUPLICATE_EXTERNAL_ID error, got %v", err)
	}

	// The old externalId (EXT123) should still be valid and findable
	retrieved, err := store.GetCustomerByExternalID(ctx, "EXT123")
	if err != nil {
		t.Fatalf("expected to find customer with old externalId EXT123, got error: %v", err)
	}
	if retrieved.ID != created1.ID {
		t.Errorf("expected to find customer with ID %d, got %d", created1.ID, retrieved.ID)
	}

	// The existing externalId (EXT456) should still belong to customer2
	retrieved2, err := store.GetCustomerByExternalID(ctx, "EXT456")
	if err != nil {
		t.Fatalf("expected to find customer with externalId EXT456, got error: %v", err)
	}
	if retrieved2.ID != created2.ID {
		t.Errorf("expected EXT456 to belong to customer %d, got %d", created2.ID, retrieved2.ID)
	}
}

func TestQueryCustomersFiltersByExternalID(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	firstCustomer := &models.Customer{
		FirstName:  "Jim",
		LastName:   "Doe",
		ExternalID: "JIM_EXT_1",
		Subsidiary: 1,
	}
	createdFirstCustomer, err := store.CreateCustomer(ctx, firstCustomer)
	if err != nil {
		t.Fatalf("failed to create first customer: %v", err)
	}

	secondCustomer := &models.Customer{
		FirstName:  "Jim",
		LastName:   "Doe",
		ExternalID: "JIM_EXT_2",
		Subsidiary: 1,
	}
	createdSecondCustomer, err := store.CreateCustomer(ctx, secondCustomer)
	if err != nil {
		t.Fatalf("failed to create second customer: %v", err)
	}

	results, err := store.QueryCustomers(ctx, []*QueryFilter{{
		Field:    "externalId",
		Operator: "=",
		Values:   []any{"JIM_EXT_2"},
	}})
	if err != nil {
		t.Fatalf("expected query to succeed, got error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected exactly 1 result, got %d", len(results))
	}

	if results[0].ID != createdSecondCustomer.ID {
		t.Fatalf("expected customer ID %d, got %d", createdSecondCustomer.ID, results[0].ID)
	}

	if results[0].ID == createdFirstCustomer.ID {
		t.Fatalf("query should not return customer ID %d", createdFirstCustomer.ID)
	}
}
