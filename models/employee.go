package models

import "time"

// Employee represents a NetSuite employee record
type Employee struct {
	ID                  int64     `json:"id,string"`
	FirstName           string    `json:"firstName"`
	LastName            string    `json:"lastName"`
	HouseholdMinistryID string    `json:"householdMinistryId,omitempty"`
	Subsidiary          int64     `json:"subsidiary"`
	CreatedAt           time.Time `json:"-"`
}

// NewEmployee creates a new employee with default values
func NewEmployee() *Employee {
	return &Employee{
		CreatedAt: time.Now().UTC(),
	}
}
