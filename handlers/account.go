package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/sil-org/netsuite-mock/domain"
	"github.com/sil-org/netsuite-mock/models"
)

// CreateAccountRequest represents an account creation request
type CreateAccountRequest struct {
	ExternalID    string `json:"externalId,omitempty"`
	AccountNumber string `json:"acctNumber,omitempty"`
	Class         string `json:"class,omitempty"`
}

// CreateAccount handles POST /services/rest/record/v1/account
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Failed to parse request body")
		return
	}

	// Check external ID uniqueness
	if req.ExternalID != "" {
		exists, err := h.store.ExternalIDExists(r.Context(), "account", req.ExternalID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError,
				"Failed to check externalId uniqueness")
			return
		}
		if exists {
			respondError(w, http.StatusConflict, domain.ErrorCodeDuplicateExternalID,
				"A record with the specified externalId already exists")
			return
		}
	}

	// Create account
	account := models.NewAccount()
	account.AccountNumber = req.AccountNumber
	account.Class = req.Class
	account.ExternalID = req.ExternalID

	created, err := h.store.CreateAccount(r.Context(), account)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "DUPLICATE_ACCOUNT_NUMBER" {
			respondError(w, http.StatusConflict, "DUPLICATE_ACCOUNT_NUMBER",
				"A record with the specified account number already exists")
		} else {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, errMsg)
		}
		return
	}

	log.Println("account created, id=", created.ID)
	respondNoContent(w, r, strconv.FormatInt(created.ID, 10))
}

// GetAccount handles GET /services/rest/record/v1/account/{id}
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Account ID is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Invalid account ID")
		return
	}

	account, err := h.store.GetAccount(r.Context(), id)
	if err != nil {
		if err.Error() == domain.ErrorCodeRecordNotFound {
			respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound, "Account not found")
		} else {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	log.Println("account found, id=", account.ID)
	respondSuccess(w, http.StatusOK, account)
}
