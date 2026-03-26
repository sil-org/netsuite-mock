package models

import "time"

// Invoice represents a NetSuite invoice record
type Invoice struct {
	ID          int64         `json:"id,string"`
	ExternalID  string        `json:"externalId,omitempty"`
	TranID      string        `json:"tranId,omitempty"`
	Entity      Entity        `json:"entity"`
	TranDate    string        `json:"tranDate"`
	Subsidiary  Subsidiary    `json:"subsidiary"`
	Currency    Currency      `json:"currency"`
	Memo        string        `json:"memo,omitempty"`
	Items       []InvoiceItem `json:"item,omitempty"`
	ShipAddress string        `json:"shipAddress,omitempty"`
	CreatedAt   time.Time     `json:"-"`
}

type Entity struct {
	ID int64 `json:"id,string"`
}

// InvoiceItem represents a line item on an invoice
type InvoiceItem struct {
	Item                   int64          `json:"item,string"`
	Amount                 float64        `json:"amount"`
	Description            string         `json:"description,omitempty"`
	IsTaxable              bool           `json:"isTaxable"`
	CustcolNpoSegmentCode  string         `json:"custcol_npo_segment_code,omitempty"`
	CustcolJournalLineDate string         `json:"custcol_journal_line_date,omitempty"`
	CustomFields           map[string]any `json:"-"`
}

// NewInvoice creates a new invoice with default values
func NewInvoice() *Invoice {
	return &Invoice{
		CreatedAt: time.Now().UTC(),
	}
}
