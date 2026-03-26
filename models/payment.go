package models

import "time"

// CustomerPayment represents a NetSuite customer payment record
type CustomerPayment struct {
	ID         int64      `json:"id,string"`
	ExternalID string     `json:"externalId,omitempty"`
	TranID     string     `json:"tranId,omitempty"`
	Customer   Customer   `json:"customer"`
	Account    Account    `json:"account"`
	AutoApply  bool       `json:"autoApply"`
	Payment    float64    `json:"payment"`
	TranDate   string     `json:"tranDate"`
	Memo       string     `json:"memo,omitempty"`
	Subsidiary Subsidiary `json:"subsidiary"`
	Class      Class      `json:"class,omitempty"`
	Currency   Currency   `json:"currency"`
	CreatedAt  time.Time  `json:"-"`
}

type Class struct {
	ID int64 `json:"id,string"`
}

type Currency struct {
	ID int64 `json:"id,string"`
}

type Subsidiary struct {
	ID int64 `json:"id,string"`
}

// NewCustomerPayment creates a new customer payment with default values
func NewCustomerPayment() *CustomerPayment {
	return &CustomerPayment{
		CreatedAt: time.Now().UTC(),
	}
}
