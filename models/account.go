package models

import "time"

// Account represents a NetSuite account record
type Account struct {
	ID            int64     `json:"id,string"`
	ExternalID    string    `json:"externalId,omitempty"`
	AccountNumber string    `json:"acctNumber"`
	Class         string    `json:"class"`
	CreatedAt     time.Time `json:"-"`
}

// NewAccount creates a new account with default values
func NewAccount() *Account {
	return &Account{
		CreatedAt: time.Now().UTC(),
	}
}
