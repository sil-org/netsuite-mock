package storage

import (
	"context"
	"errors"

	"github.com/sil-org/netsuite-mock/models"
)

var ErrInvalidQueryFilter = errors.New("INVALID_QUERY_FILTER")

// QueryFilter represents filtering parameters for queries
type QueryFilter struct {
	Field    string
	Operator string
	Values   []any
}

// Store defines all storage operations
type Store interface {
	// Customer operations
	CreateCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error)
	GetCustomer(ctx context.Context, id int64) (*models.Customer, error)
	GetCustomerByExternalID(ctx context.Context, externalID string) (*models.Customer, error)
	UpdateCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error)
	QueryCustomers(ctx context.Context, filters []*QueryFilter) ([]*models.Customer, error)

	// Employee operations
	CreateEmployee(ctx context.Context, employee *models.Employee) (*models.Employee, error)
	GetEmployee(ctx context.Context, id int64) (*models.Employee, error)
	QueryEmployees(ctx context.Context, filters []*QueryFilter) ([]*models.Employee, error)

	// Account operations
	CreateAccount(ctx context.Context, account *models.Account) (*models.Account, error)
	GetAccount(ctx context.Context, id int64) (*models.Account, error)
	GetAccountByNumber(ctx context.Context, number string) (*models.Account, error)
	QueryAccounts(ctx context.Context, filters []*QueryFilter) ([]*models.Account, error)

	// Invoice operations
	CreateInvoice(ctx context.Context, invoice *models.Invoice) (*models.Invoice, error)
	GetInvoice(ctx context.Context, id int64) (*models.Invoice, error)

	// Customer Payment operations
	CreateCustomerPayment(ctx context.Context, payment *models.CustomerPayment) (*models.CustomerPayment, error)
	GetCustomerPayment(ctx context.Context, id int64) (*models.CustomerPayment, error)

	// Journal Entry operations
	CreateJournalEntry(ctx context.Context, entry *models.JournalEntry) (*models.JournalEntry, error)
	GetJournalEntry(ctx context.Context, id int64) (*models.JournalEntry, error)

	// Utility
	ExternalIDExists(ctx context.Context, recordType string, externalID string) (bool, error)
	Close(ctx context.Context) error
}
