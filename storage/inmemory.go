package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/sil-org/netsuite-mock/domain"
	"github.com/sil-org/netsuite-mock/models"
)

// InMemoryStore implements the Store interface using in-memory maps
type InMemoryStore struct {
	mu               sync.RWMutex
	customers        map[int64]*models.Customer
	employees        map[int64]*models.Employee
	accounts         map[int64]*models.Account
	invoices         map[int64]*models.Invoice
	payments         map[int64]*models.CustomerPayment
	journalEntries   map[int64]*models.JournalEntry
	externalIDIndex  map[string]map[string]int64 // recordType -> externalID -> id
	accountNumberIdx map[string]int64            // accountNumber -> id
	nextID           map[string]int64            // recordType -> nextID
}

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		customers:        make(map[int64]*models.Customer),
		employees:        make(map[int64]*models.Employee),
		accounts:         make(map[int64]*models.Account),
		invoices:         make(map[int64]*models.Invoice),
		payments:         make(map[int64]*models.CustomerPayment),
		journalEntries:   make(map[int64]*models.JournalEntry),
		externalIDIndex:  make(map[string]map[string]int64),
		accountNumberIdx: make(map[string]int64),
		nextID: map[string]int64{
			"customer": 1,
			"employee": 1,
			"account":  1,
			"invoice":  1,
			"payment":  1,
			"journal":  1,
		},
	}
}

// ExternalIDExists checks if an externalId already exists for a record type
func (s *InMemoryStore) ExternalIDExists(ctx context.Context, recordType string, externalID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if externalID == "" {
		return false, nil // Empty externalIDs don't need to be unique
	}

	idx, ok := s.externalIDIndex[recordType]
	if !ok {
		return false, nil
	}

	_, exists := idx[externalID]
	return exists, nil
}

// nextIDForType generates the next ID for a record type
func (s *InMemoryStore) nextIDForType(recordType string) int64 {
	id := s.nextID[recordType]
	s.nextID[recordType] = id + 1
	return id
}

// indexExternalID adds an externalID to the index
func (s *InMemoryStore) indexExternalID(recordType string, externalID string, id int64) {
	if externalID == "" {
		return // Don't index empty externalIDs
	}

	if _, ok := s.externalIDIndex[recordType]; !ok {
		s.externalIDIndex[recordType] = make(map[string]int64)
	}
	s.externalIDIndex[recordType][externalID] = id
}

// removeExternalIDIndex removes an externalID from the index
func (s *InMemoryStore) removeExternalIDIndex(recordType string, externalID string) {
	if externalID == "" {
		return
	}

	if idx, ok := s.externalIDIndex[recordType]; ok {
		delete(idx, externalID)
	}
}

// ============================================================================
// Customer Operations
// ============================================================================

// CreateCustomer creates a new customer record
func (s *InMemoryStore) CreateCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check externalID uniqueness if provided
	if customer.ExternalID != "" {
		if idx, ok := s.externalIDIndex["customer"]; ok {
			if _, exists := idx[customer.ExternalID]; exists {
				return nil, fmt.Errorf(domain.ErrorCodeDuplicateExternalID)
			}
		}
	}

	// Assign ID
	customer.ID = s.nextIDForType("customer")

	// Store customer
	s.customers[customer.ID] = customer

	// Index externalID if provided
	s.indexExternalID("customer", customer.ExternalID, customer.ID)

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (s *InMemoryStore) GetCustomer(ctx context.Context, id int64) (*models.Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	customer, ok := s.customers[id]
	if !ok {
		return nil, fmt.Errorf(domain.ErrorCodeRecordNotFound)
	}

	return customer, nil
}

// GetCustomerByExternalID retrieves a customer by externalId
func (s *InMemoryStore) GetCustomerByExternalID(ctx context.Context, externalID string) (*models.Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idx, ok := s.externalIDIndex["customer"]
	if !ok {
		return nil, fmt.Errorf(domain.ErrorCodeRecordNotFound)
	}

	id, ok := idx[externalID]
	if !ok {
		return nil, fmt.Errorf(domain.ErrorCodeRecordNotFound)
	}

	return s.customers[id], nil
}

// UpdateCustomer updates an existing customer record
func (s *InMemoryStore) UpdateCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.customers[customer.ID]
	if !ok {
		return nil, fmt.Errorf(domain.ErrorCodeRecordNotFound)
	}

	// Check if externalID needs to be reindexed
	// Note: existing and customer might be the same object, so we need to check before any modifications
	externalIDChanged := customer.ExternalID != existing.ExternalID

	if externalIDChanged {
		// Get the old external ID from what's currently stored
		oldExternalID := existing.ExternalID

		// Check new externalID uniqueness if provided BEFORE removing old one
		if customer.ExternalID != "" {
			if idx, ok := s.externalIDIndex["customer"]; ok {
				if _, exists := idx[customer.ExternalID]; exists {
					return nil, fmt.Errorf(domain.ErrorCodeDuplicateExternalID)
				}
			}
		}

		// Only remove old externalID from index after validation passes
		s.removeExternalIDIndex("customer", oldExternalID)

		// Index new externalID
		s.indexExternalID("customer", customer.ExternalID, customer.ID)
	}

	// Now update the stored customer
	s.customers[customer.ID] = customer
	return customer, nil
}

// QueryCustomers queries customers with filters
func (s *InMemoryStore) QueryCustomers(ctx context.Context, filters []*QueryFilter) ([]*models.Customer, error) {
	if err := validateFilters(filters, customerFilterFields); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*models.Customer

	for _, customer := range s.customers {
		if matchesFilters(customer, filters) {
			results = append(results, customer)
		}
	}

	return results, nil
}

// ============================================================================
// Employee Operations
// ============================================================================

// CreateEmployee creates a new employee record
func (s *InMemoryStore) CreateEmployee(ctx context.Context, employee *models.Employee) (*models.Employee, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check externalID uniqueness if provided
	if employee.HouseholdMinistryID != "" {
		if idx, ok := s.externalIDIndex["employee"]; ok {
			if _, exists := idx[employee.HouseholdMinistryID]; exists {
				return nil, fmt.Errorf(domain.ErrorCodeDuplicateExternalID)
			}
		}
	}

	// Assign ID
	employee.ID = s.nextIDForType("employee")

	// Store employee
	s.employees[employee.ID] = employee

	// Index externalID if provided
	s.indexExternalID("employee", employee.HouseholdMinistryID, employee.ID)

	return employee, nil
}

// GetEmployee retrieves an employee by ID
func (s *InMemoryStore) GetEmployee(ctx context.Context, id int64) (*models.Employee, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	employee, ok := s.employees[id]
	if !ok {
		return nil, fmt.Errorf(domain.ErrorCodeRecordNotFound)
	}

	return employee, nil
}

// QueryEmployees queries employees with filters
func (s *InMemoryStore) QueryEmployees(ctx context.Context, filters []*QueryFilter) ([]*models.Employee, error) {
	if err := validateFilters(filters, employeeFilterFields); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*models.Employee

	for _, employee := range s.employees {
		if matchesFilters(employee, filters) {
			results = append(results, employee)
		}
	}

	return results, nil
}

// ============================================================================
// Account Operations
// ============================================================================

// CreateAccount creates a new account record
func (s *InMemoryStore) CreateAccount(ctx context.Context, account *models.Account) (*models.Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check externalID uniqueness if provided
	if account.ExternalID != "" {
		if idx, ok := s.externalIDIndex["account"]; ok {
			if _, exists := idx[account.ExternalID]; exists {
				return nil, fmt.Errorf(domain.ErrorCodeDuplicateExternalID)
			}
		}
	}

	// Check account number uniqueness
	if _, exists := s.accountNumberIdx[account.AccountNumber]; exists {
		return nil, fmt.Errorf("DUPLICATE_ACCOUNT_NUMBER")
	}

	// Assign ID
	account.ID = s.nextIDForType("account")

	// Store account
	s.accounts[account.ID] = account

	// Index by account number
	s.accountNumberIdx[account.AccountNumber] = account.ID

	// Index externalID if provided
	s.indexExternalID("account", account.ExternalID, account.ID)

	return account, nil
}

// GetAccount retrieves an account by ID
func (s *InMemoryStore) GetAccount(ctx context.Context, id int64) (*models.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, ok := s.accounts[id]
	if !ok {
		return nil, fmt.Errorf(domain.ErrorCodeRecordNotFound)
	}

	return account, nil
}

// GetAccountByNumber retrieves an account by account number
func (s *InMemoryStore) GetAccountByNumber(ctx context.Context, number string) (*models.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.accountNumberIdx[number]
	if !ok {
		return nil, fmt.Errorf(domain.ErrorCodeRecordNotFound)
	}

	return s.accounts[id], nil
}

// QueryAccounts queries accounts with filters
func (s *InMemoryStore) QueryAccounts(ctx context.Context, filters []*QueryFilter) ([]*models.Account, error) {
	if err := validateFilters(filters, accountFilterFields); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*models.Account

	for _, account := range s.accounts {
		if matchesFilters(account, filters) {
			results = append(results, account)
		}
	}

	return results, nil
}

// ============================================================================
// Invoice Operations
// ============================================================================

// CreateInvoice creates a new invoice record
func (s *InMemoryStore) CreateInvoice(ctx context.Context, invoice *models.Invoice) (*models.Invoice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check externalID uniqueness if provided
	if invoice.ExternalID != "" {
		if idx, ok := s.externalIDIndex["invoice"]; ok {
			if _, exists := idx[invoice.ExternalID]; exists {
				return nil, fmt.Errorf(domain.ErrorCodeDuplicateExternalID)
			}
		}
	}

	// Verify customer exists
	if _, ok := s.customers[invoice.Entity.ID]; !ok {
		return nil, fmt.Errorf(domain.ErrorCodeRecordNotFound)
	}

	// Assign ID
	invoice.ID = s.nextIDForType("invoice")
	invoice.TranID = fmt.Sprintf("INV-%d", invoice.ID)

	// Store invoice
	s.invoices[invoice.ID] = invoice

	// Index externalID if provided
	s.indexExternalID("invoice", invoice.ExternalID, invoice.ID)

	return invoice, nil
}

// GetInvoice retrieves an invoice by ID
func (s *InMemoryStore) GetInvoice(ctx context.Context, id int64) (*models.Invoice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	invoice, ok := s.invoices[id]
	if !ok {
		return nil, fmt.Errorf(domain.ErrorCodeRecordNotFound)
	}

	return invoice, nil
}

// ============================================================================
// Customer Payment Operations
// ============================================================================

// CreateCustomerPayment creates a new customer payment record
func (s *InMemoryStore) CreateCustomerPayment(ctx context.Context, payment *models.CustomerPayment) (*models.CustomerPayment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check externalID uniqueness if provided
	if payment.ExternalID != "" {
		if idx, ok := s.externalIDIndex["payment"]; ok {
			if _, exists := idx[payment.ExternalID]; exists {
				return nil, fmt.Errorf(domain.ErrorCodeDuplicateExternalID)
			}
		}
	}

	// Verify customer exists
	if _, ok := s.customers[payment.Customer.ID]; !ok {
		return nil, fmt.Errorf(domain.ErrorCodeRecordNotFound)
	}

	// Verify account exists
	if _, ok := s.accounts[payment.Account.ID]; !ok {
		return nil, fmt.Errorf(domain.ErrorCodeRecordNotFound)
	}

	// Assign ID
	payment.ID = s.nextIDForType("payment")
	payment.TranID = fmt.Sprintf("CUST-PAY-%d", payment.ID)

	// Store payment
	s.payments[payment.ID] = payment

	// Index externalID if provided
	s.indexExternalID("payment", payment.ExternalID, payment.ID)

	return payment, nil
}

// GetCustomerPayment retrieves a customer payment by ID
func (s *InMemoryStore) GetCustomerPayment(ctx context.Context, id int64) (*models.CustomerPayment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	payment, ok := s.payments[id]
	if !ok {
		return nil, fmt.Errorf(domain.ErrorCodeRecordNotFound)
	}

	return payment, nil
}

// ============================================================================
// Journal Entry Operations
// ============================================================================

// CreateJournalEntry creates a new journal entry record
func (s *InMemoryStore) CreateJournalEntry(ctx context.Context, entry *models.JournalEntry) (*models.JournalEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check externalID uniqueness if provided
	if entry.ExternalID != "" {
		if idx, ok := s.externalIDIndex["journal"]; ok {
			if _, exists := idx[entry.ExternalID]; exists {
				return nil, fmt.Errorf(domain.ErrorCodeDuplicateExternalID)
			}
		}
	}

	// Validate at least 2 lines
	if len(entry.Lines) < 2 {
		return nil, fmt.Errorf(domain.ErrorCodeInvalidRequest)
	}

	// Validate debits equal credits
	var totalDebits, totalCredits float64
	for _, line := range entry.Lines {
		totalDebits += line.Debit
		totalCredits += line.Credit
	}

	// Allow small floating point differences
	if totalDebits-totalCredits > 0.01 || totalCredits-totalDebits > 0.01 {
		return nil, fmt.Errorf(domain.ErrorCodeDebitCreditMismatch)
	}

	// Assign ID
	entry.ID = s.nextIDForType("journal")
	entry.TranID = fmt.Sprintf("JE-%d", entry.ID)

	// Store entry
	s.journalEntries[entry.ID] = entry

	// Index externalID if provided
	s.indexExternalID("journal", entry.ExternalID, entry.ID)

	return entry, nil
}

// GetJournalEntry retrieves a journal entry by ID
func (s *InMemoryStore) GetJournalEntry(ctx context.Context, id int64) (*models.JournalEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.journalEntries[id]
	if !ok {
		return nil, fmt.Errorf(domain.ErrorCodeRecordNotFound)
	}

	return entry, nil
}

// ============================================================================
// Utility Methods
// ============================================================================

// Close closes the store (no-op for in-memory store)
func (s *InMemoryStore) Close(ctx context.Context) error {
	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

var customerFilterFields = map[string]struct{}{
	"id":         {},
	"externalid": {},
	"firstname":  {},
	"lastname":   {},
	"subsidiary": {},
	"isinactive": {},
}

var employeeFilterFields = map[string]struct{}{
	"id":                             {},
	"externalid":                     {},
	"firstname":                      {},
	"lastname":                       {},
	"custentity_nscs_hh_ministry_id": {},
}

var accountFilterFields = map[string]struct{}{
	"id":         {},
	"externalid": {},
	"acctnumber": {},
	"class":      {},
}

func validateFilters(filters []*QueryFilter, allowedFields map[string]struct{}) error {
	for _, filter := range filters {
		if filter == nil {
			return fmt.Errorf("filter cannot be nil: %w", ErrInvalidQueryFilter)
		}

		normalizedField := strings.ToLower(strings.TrimSpace(filter.Field))
		if normalizedField == "" {
			return fmt.Errorf("filter field is required: %w", ErrInvalidQueryFilter)
		}

		normalizedOperator := strings.TrimSpace(filter.Operator)
		if normalizedOperator != "=" {
			return fmt.Errorf("operator %q is not supported (only \"=\"): %w", normalizedOperator, ErrInvalidQueryFilter)
		}

		if len(filter.Values) == 0 {
			return fmt.Errorf("filter values are required for field %q: %w", filter.Field, ErrInvalidQueryFilter)
		}

		if _, ok := allowedFields[normalizedField]; !ok {
			return fmt.Errorf("field %q is not recognized: %w", filter.Field, ErrInvalidQueryFilter)
		}

		filter.Field = normalizedField
		filter.Operator = normalizedOperator
	}

	return nil
}

// matchesFilters checks if a record matches all provided filters
func matchesFilters(record any, filters []*QueryFilter) bool {
	if len(filters) == 0 {
		return true
	}

	for _, filter := range filters {
		if filter == nil {
			return false
		}

		if !recordMatchesFilter(record, strings.ToLower(strings.TrimSpace(filter.Field)), filter.Values[0]) {
			return false
		}
	}

	return true
}

func recordMatchesFilter(record any, field string, filterValue any) bool {
	switch typedRecord := record.(type) {
	case *models.Customer:
		return customerMatchesFilter(typedRecord, field, filterValue)
	case *models.Employee:
		return employeeMatchesFilter(typedRecord, field, filterValue)
	case *models.Account:
		return accountMatchesFilter(typedRecord, field, filterValue)
	default:
		return false
	}
}

func customerMatchesFilter(customer *models.Customer, field string, filterValue any) bool {
	switch field {
	case "id":
		return matchesInt64(customer.ID, filterValue)
	case "externalid":
		return matchesString(customer.ExternalID, filterValue)
	case "firstname":
		return matchesString(customer.FirstName, filterValue)
	case "lastname":
		return matchesString(customer.LastName, filterValue)
	case "subsidiary":
		return matchesInt64(customer.Subsidiary, filterValue)
	case "isinactive":
		return matchesInactive(customer.IsInactive, filterValue)
	default:
		return false
	}
}

func employeeMatchesFilter(employee *models.Employee, field string, filterValue any) bool {
	switch field {
	case "id":
		return matchesInt64(employee.ID, filterValue)
	case "firstname":
		return matchesString(employee.FirstName, filterValue)
	case "lastname":
		return matchesString(employee.LastName, filterValue)
	case "custentity_nscs_hh_ministry_id":
		return matchesString(employee.HouseholdMinistryID, filterValue)
	default:
		return false
	}
}

func accountMatchesFilter(account *models.Account, field string, filterValue any) bool {
	switch field {
	case "id":
		return matchesInt64(account.ID, filterValue)
	case "externalid":
		return matchesString(account.ExternalID, filterValue)
	case "acctnumber":
		return matchesString(account.AccountNumber, filterValue)
	case "class":
		return matchesString(account.Class, filterValue)
	default:
		return false
	}
}

func matchesString(actualValue string, filterValue any) bool {
	return actualValue == strings.TrimSpace(fmt.Sprintf("%v", filterValue))
}

func matchesInt64(actualValue int64, filterValue any) bool {
	switch typedValue := filterValue.(type) {
	case int:
		return actualValue == int64(typedValue)
	case int32:
		return actualValue == int64(typedValue)
	case int64:
		return actualValue == typedValue
	case float64:
		return actualValue == int64(typedValue)
	default:
		valueString := strings.TrimSpace(fmt.Sprintf("%v", filterValue))
		parsedValue, parseErr := strconv.ParseInt(valueString, 10, 64)
		if parseErr != nil {
			return false
		}
		return actualValue == parsedValue
	}
}

func matchesInactive(actualValue bool, filterValue any) bool {
	switch typedValue := filterValue.(type) {
	case bool:
		return actualValue == typedValue
	default:
		normalizedValue := strings.ToUpper(strings.TrimSpace(fmt.Sprintf("%v", filterValue)))
		switch normalizedValue {
		case "T", "TRUE":
			return actualValue
		case "F", "FALSE":
			return !actualValue
		default:
			return false
		}
	}
}
