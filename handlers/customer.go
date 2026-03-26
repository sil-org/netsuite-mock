package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/sil-org/netsuite-mock/domain"
	"github.com/sil-org/netsuite-mock/models"
)

// CreateCustomerRequest represents a customer creation request
type CreateCustomerRequest struct {
	FirstName                 string              `json:"firstName"`
	LastName                  string              `json:"lastName"`
	ExternalID                string              `json:"externalId,omitempty"`
	Subsidiary                Related             `json:"subsidiary"`
	IsInactive                bool                `json:"isInactive"`
	CustentitySilCustCategory int                 `json:"custentity_sil_cust_category,omitempty"`
	CustentityCustFieldSil    int                 `json:"custentity_cust_field_sil,omitempty"`
	AddressBook               *models.AddressBook `json:"addressBook,omitempty"`
}

// PatchCustomerRequest represents a customer update request
type PatchCustomerRequest struct {
	FirstName  string `json:"firstName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	IsInactive bool   `json:"isInactive"`
}

// CreateCustomer handles POST /services/rest/record/v1/customer
func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var req CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Failed to parse request body")
		return
	}

	// Validate required fields
	if req.FirstName == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeMissingRequiredField, "firstName is required")
		return
	}
	if req.LastName == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeMissingRequiredField, "lastName is required")
		return
	}

	// Check external ID uniqueness
	if req.ExternalID != "" {
		exists, err := h.store.ExternalIDExists(r.Context(), "customer", req.ExternalID)
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

	// Create customer
	customer := models.NewCustomer()
	customer.FirstName = req.FirstName
	customer.LastName = req.LastName
	customer.ExternalID = req.ExternalID
	customer.Subsidiary = req.Subsidiary.ID
	customer.IsInactive = req.IsInactive
	customer.AddressBook = req.AddressBook

	created, err := h.store.CreateCustomer(r.Context(), customer)
	if err != nil {
		respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		return
	}

	log.Println("customer created, id=", created.ID)
	respondNoContent(w, r, strconv.FormatInt(created.ID, 10))
}

// GetCustomer handles GET /services/rest/record/v1/customer/{id}
func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Customer ID is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Invalid customer ID")
		return
	}

	customer, err := h.store.GetCustomer(r.Context(), id)
	if err != nil {
		if err.Error() == domain.ErrorCodeRecordNotFound {
			respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound, "Customer not found")
		} else {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	log.Println("customer found, id=", customer.ID)
	respondSuccess(w, http.StatusOK, customer)
}

// UpdateCustomer handles PATCH /services/rest/record/v1/customer/{id}
func (h *Handler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Customer ID is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Invalid customer ID")
		return
	}

	var req PatchCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Failed to parse request body")
		return
	}

	// Get existing customer
	customer, err := h.store.GetCustomer(r.Context(), id)
	if err != nil {
		if err.Error() == domain.ErrorCodeRecordNotFound {
			respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound, "Customer not found")
		} else {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	// Update fields
	if req.FirstName != "" {
		customer.FirstName = req.FirstName
	}
	if req.LastName != "" {
		customer.LastName = req.LastName
	}
	customer.IsInactive = req.IsInactive

	// Update in store
	updated, err := h.store.UpdateCustomer(r.Context(), customer)
	if err != nil {
		if err.Error() == domain.ErrorCodeDuplicateExternalID {
			respondError(w, http.StatusConflict, domain.ErrorCodeDuplicateExternalID,
				"A record with the specified externalId already exists")
		} else {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	log.Println("customer updated, id=", updated.ID)
	respondNoContent(w, r, strconv.FormatInt(updated.ID, 10))
}
