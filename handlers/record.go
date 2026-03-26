package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sil-org/netsuite-mock/domain"
	"github.com/sil-org/netsuite-mock/models"
)

// CreateInvoiceRequest represents an invoice creation request
type CreateInvoiceRequest struct {
	Entity      Related              `json:"entity"`
	Date        string               `json:"date"`
	Subsidiary  Related              `json:"subsidiary"`
	Currency    Related              `json:"currency"`
	Memo        string               `json:"memo,omitempty"`
	Item        *InvoiceItemsWrapper `json:"item,omitempty"`
	ExternalID  string               `json:"externalId,omitempty"`
	ShipAddress string               `json:"shipAddress,omitempty"`
	TranDate    string               `json:"tranDate,omitempty"`
}

// InvoiceItemsWrapper wraps invoice items for the API
type InvoiceItemsWrapper struct {
	Items []InvoiceItemRequest `json:"items"`
}

// InvoiceItem represents a line item for creation of an invoice
type InvoiceItemRequest struct {
	Item                   Related `json:"item"`
	Amount                 float64 `json:"amount"`
	Description            string  `json:"description,omitempty"`
	IsTaxable              bool    `json:"isTaxable"`
	CustcolNpoSegmentCode  string  `json:"custcol_npo_segment_code,omitempty"`
	CustcolJournalLineDate string  `json:"custcol_journal_line_date,omitempty"`
}

// CreateInvoice handles POST /services/rest/record/v1/invoice
func (h *Handler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Failed to parse request body"+err.Error())
		return
	}

	// Check external ID uniqueness
	if req.ExternalID != "" {
		exists, err := h.store.ExternalIDExists(r.Context(), "invoice", req.ExternalID)
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

	// Create invoice
	invoice := models.NewInvoice()
	invoice.Currency = models.Currency{ID: req.Currency.ID}
	invoice.Entity = models.Entity{ID: req.Entity.ID}
	invoice.ExternalID = req.ExternalID
	invoice.Memo = req.Memo
	invoice.ShipAddress = req.ShipAddress
	invoice.Subsidiary = models.Subsidiary{ID: req.Subsidiary.ID}
	invoice.TranDate = req.TranDate
	if req.Item != nil {
		items, err := parseInvoiceItem(req)
		if err != nil {
			respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidFieldValue, "Item parse error: "+err.Error())
			return
		}
		invoice.Items = items
	}

	created, err := h.store.CreateInvoice(r.Context(), invoice)
	if err != nil {
		if err.Error() == domain.ErrorCodeRecordNotFound {
			respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound, "Referenced customer or subsidiary not found")
		} else {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	respondNoContent(w, r, strconv.FormatInt(created.ID, 10))
}

func parseInvoiceItem(req CreateInvoiceRequest) ([]models.InvoiceItem, error) {
	items := make([]models.InvoiceItem, len(req.Item.Items))
	for i, item := range req.Item.Items {
		items[i].Amount = item.Amount
		items[i].CustcolJournalLineDate = item.CustcolJournalLineDate
		items[i].CustcolNpoSegmentCode = item.CustcolNpoSegmentCode
		items[i].Description = item.Description
		items[i].IsTaxable = item.IsTaxable
		items[i].Item = item.Item.ID
	}
	return items, nil
}

// InvoiceItemsResponse represents the NetSuite sublist response for invoice items
type InvoiceItemsResponse struct {
	Count        int                  `json:"count"`
	HasMore      bool                 `json:"hasMore"`
	Items        []models.InvoiceItem `json:"items"`
	Offset       int                  `json:"offset"`
	TotalResults int                  `json:"totalResults"`
}

// GetInvoiceItem handles GET /services/rest/record/v1/invoice/{id}/item/{lineNum}
// lineNum is 1-based
func (h *Handler) GetInvoiceItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Invoice ID is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Invalid invoice ID")
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

	invoice, err := h.store.GetInvoice(r.Context(), id)
	if err != nil {
		if err.Error() == domain.ErrorCodeRecordNotFound {
			respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound, "Invoice not found")
		} else {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	if lineNum > len(invoice.Items) || lineNum <= 0 {
		respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound, fmt.Sprintf("Invoice item at line %d not found", lineNum))
		return
	}

	respondSuccess(w, http.StatusOK, invoice.Items[lineNum-1])
}

// GetInvoice handles GET /services/rest/record/v1/invoice/{id}
func (h *Handler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Invoice ID is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Invalid invoice ID")
		return
	}

	invoice, err := h.store.GetInvoice(r.Context(), id)
	if err != nil {
		if err.Error() == domain.ErrorCodeRecordNotFound {
			respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound, "Invoice not found")
		} else {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	respondSuccess(w, http.StatusOK, invoice)
}

// CreatePaymentRequest represents a customer payment creation request
type CreatePaymentRequest struct {
	Customer   Related `json:"customer"`
	Account    Related `json:"account"`
	AutoApply  bool    `json:"autoApply"`
	Payment    float64 `json:"payment"`
	TranDate   string  `json:"tranDate"`
	Memo       string  `json:"memo,omitempty"`
	Subsidiary Related `json:"subsidiary"`
	Class      Related `json:"class,omitempty"`
	Currency   Related `json:"currency"`
	ExternalID string  `json:"externalId,omitempty"`
}

// CreatePayment handles POST /services/rest/record/v1/customerPayment
func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Failed to parse request body")
		return
	}

	// Check external ID uniqueness
	if req.ExternalID != "" {
		exists, err := h.store.ExternalIDExists(r.Context(), "payment", req.ExternalID)
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

	// Create payment
	payment := models.NewCustomerPayment()
	payment.Customer = models.Customer{ID: req.Customer.ID}
	payment.Account = models.Account{ID: req.Account.ID}
	payment.AutoApply = req.AutoApply
	payment.Payment = req.Payment
	payment.Subsidiary = models.Subsidiary{ID: req.Subsidiary.ID}
	payment.Class = models.Class{ID: req.Class.ID}
	payment.Currency = models.Currency{ID: req.Currency.ID}
	payment.Memo = req.Memo
	payment.ExternalID = req.ExternalID
	payment.TranDate = req.TranDate

	created, err := h.store.CreateCustomerPayment(r.Context(), payment)
	if err != nil {
		if err.Error() == domain.ErrorCodeRecordNotFound {
			respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound, "Referenced customer or account not found")
		} else {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	respondNoContent(w, r, strconv.FormatInt(created.ID, 10))
}

// GetPayment handles GET /services/rest/record/v1/customerPayment/{id}
func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Payment ID is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Invalid payment ID")
		return
	}

	payment, err := h.store.GetCustomerPayment(r.Context(), id)
	if err != nil {
		if err.Error() == domain.ErrorCodeRecordNotFound {
			respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound, "Payment not found")
		} else {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	respondSuccess(w, http.StatusOK, payment)
}
