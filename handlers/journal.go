package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sil-org/netsuite-mock/domain"
	"github.com/sil-org/netsuite-mock/models"
)

// CreateJournalEntryRequest represents a journal entry creation request
type CreateJournalEntryRequest struct {
	Memo       string               `json:"memo,omitempty"`
	Subsidiary Related              `json:"subsidiary"`
	Line       *JournalLinesWrapper `json:"line"`
	ExternalID string               `json:"externalId,omitempty"`
	TranDate   string               `json:"tranDate,omitempty"`
}

type JournalLinesWrapper struct {
	Items []JournalLine `json:"items"`
}

type JournalLine struct {
	Account                Related `json:"account"`
	Credit                 float64 `json:"credit"`
	Debit                  float64 `json:"debit"`
	Entity                 Related `json:"entity"`
	Memo                   string  `json:"memo,omitempty"`
	CustcolJournalLineDate string  `json:"custcol_journal_line_date,omitempty"`
	Cseg3                  string  `json:"cseg3,omitempty"`
}

// CreateJournalEntry handles POST /services/rest/record/v1/journalEntry
func (h *Handler) CreateJournalEntry(w http.ResponseWriter, r *http.Request) {
	var req CreateJournalEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Failed to parse request body: "+err.Error())
		return
	}

	// Validate memo
	if req.Memo == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeMissingRequiredField, "memo is required")
		return
	}

	// Check external ID uniqueness
	if req.ExternalID != "" {
		exists, err := h.store.ExternalIDExists(r.Context(), "journal", req.ExternalID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, "Failed to check externalId uniqueness")
			return
		}
		if exists {
			respondError(w, http.StatusConflict, domain.ErrorCodeDuplicateExternalID,
				"A record with the specified externalId already exists")
			return
		}
	}

	// Create journal entry
	entry := models.NewJournalEntry()
	entry.Memo = req.Memo
	entry.Subsidiary.ID = req.Subsidiary.ID
	entry.TranDate = req.TranDate
	entry.ExternalID = req.ExternalID
	if req.Line != nil {
		items, err := parseJournalLine(req)
		if err != nil {
			respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidFieldValue, "Item parse error: "+err.Error())
			return
		}
		entry.Lines = items
	}

	created, err := h.store.CreateJournalEntry(r.Context(), entry)
	if err != nil {
		switch err.Error() {
		case domain.ErrorCodeRecordNotFound:
			respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound, "Referenced account not found")
		case domain.ErrorCodeInvalidRequest:
			respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Journal entry must have at least 2 lines")
		case domain.ErrorCodeDebitCreditMismatch:
			respondError(w, http.StatusBadRequest, domain.ErrorCodeDebitCreditMismatch, "Total debits must equal total credits")
		default:
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	respondNoContent(w, r, strconv.FormatInt(created.ID, 10))
}

func parseJournalLine(req CreateJournalEntryRequest) ([]models.JournalLine, error) {
	lines := make([]models.JournalLine, len(req.Line.Items))
	for i, line := range req.Line.Items {
		lines[i].Account.ID = line.Account.ID
		lines[i].Entity.ID = line.Entity.ID
		lines[i].Credit = line.Credit
		lines[i].Debit = line.Debit
		lines[i].CustcolJournalLineDate = line.CustcolJournalLineDate
		lines[i].Cseg3 = line.Cseg3
		lines[i].Memo = line.Memo
	}
	return lines, nil
}

// GetJournalEntry handles GET /services/rest/record/v1/journalEntry/{id}
func (h *Handler) GetJournalEntry(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Journal entry ID is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Invalid journal entry ID")
		return
	}

	entry, err := h.store.GetJournalEntry(r.Context(), id)
	if err != nil {
		if err.Error() == domain.ErrorCodeRecordNotFound {
			respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound, "Journal entry not found")
		} else {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	respondSuccess(w, http.StatusOK, entry)
}

// GetJournalEntryLine handles GET /services/rest/record/v1/journalEntry/{id}/line/{lineNum}
// lineNum is 0-based (unlike Invoice items, which are 1-based).
func (h *Handler) GetJournalEntryLine(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "JournalEntry ID is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Invalid journalEntry ID")
		return
	}

	lineStr := r.PathValue("lineNum")
	if lineStr == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Line number is required")
		return
	}

	lineNum, err := strconv.Atoi(lineStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Invalid line number; must be a positive integer")
		return
	}

	journalEntry, err := h.store.GetJournalEntry(r.Context(), id)
	if err != nil {
		if err.Error() == domain.ErrorCodeRecordNotFound {
			respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound, "JournalEntry not found")
		} else {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	if lineNum >= len(journalEntry.Lines) || lineNum < 0 {
		respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound,
			fmt.Sprintf("JournalEntry line at line %d not found", lineNum))
		return
	}

	respondSuccess(w, http.StatusOK, journalEntry.Lines[lineNum])
}
