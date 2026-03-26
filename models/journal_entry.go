package models

import "time"

// JournalEntry represents a NetSuite journal entry record
type JournalEntry struct {
	ID         int64         `json:"id,string"`
	ExternalID string        `json:"externalId,omitempty"`
	TranID     string        `json:"tranId,omitempty"`
	Memo       string        `json:"memo"`
	Subsidiary Subsidiary    `json:"subsidiary"`
	TranDate   string        `json:"tranDate,omitempty"`
	Lines      []JournalLine `json:"line,omitempty"`
	CreatedAt  time.Time     `json:"-"`
}

// JournalLine represents a single line item in a journal entry
type JournalLine struct {
	Account                Account        `json:"account"`
	AccountExternalID      string         `json:"accountExternalId,omitempty"`
	Credit                 float64        `json:"credit"`
	Debit                  float64        `json:"debit"`
	Entity                 Entity         `json:"entity"`
	Memo                   string         `json:"memo,omitempty"`
	CustcolJournalLineDate string         `json:"custcol_journal_line_date,omitempty"`
	Cseg3                  string         `json:"cseg3,omitempty"`
	CustomFields           map[string]any `json:"-"`
}

// JournalLineAccount represents account information that can include externalId
type JournalLineAccount struct {
	ID         int64  `json:"id,omitempty"`
	ExternalID string `json:"externalId,omitempty"`
}

// JournalLinesWrapper wraps journal lines for the API
type JournalLinesWrapper struct {
	Lines []JournalLine `json:"lines"`
}

// NewJournalEntry creates a new journal entry with default values
func NewJournalEntry() *JournalEntry {
	return &JournalEntry{
		CreatedAt: time.Now().UTC(),
		Lines:     []JournalLine{},
	}
}
