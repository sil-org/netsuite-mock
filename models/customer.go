package models

import "time"

// Customer represents a NetSuite customer record
type Customer struct {
	ID           int64          `json:"id,string"`
	ExternalID   string         `json:"externalId,omitempty"`
	FirstName    string         `json:"firstName"`
	LastName     string         `json:"lastName"`
	Subsidiary   int64          `json:"subsidiary,string"`
	IsInactive   bool           `json:"isInactive"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
	CustomFields map[string]any `json:"-"`
	AddressBook  *AddressBook   `json:"addressBook,omitempty"`
}

// AddressBook represents an address book entry
type AddressBook struct {
	Items []AddressBookItem `json:"items,omitempty"`
}

// AddressBookItem represents a single address book item
type AddressBookItem struct {
	AddressBookAddress *AddressBookAddress `json:"addressBookAddress,omitempty"`
}

// AddressBookAddress represents an address
type AddressBookAddress struct {
	Country string `json:"country,omitempty"`
}

// NewCustomer creates a new customer with default values
func NewCustomer() *Customer {
	return &Customer{
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		CustomFields: make(map[string]any),
	}
}
