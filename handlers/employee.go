package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/sil-org/netsuite-mock/domain"
	"github.com/sil-org/netsuite-mock/models"
)

// CreateEmployeeRequest represents a employee creation request
type CreateEmployeeRequest struct {
	FirstName           string  `json:"firstName"`
	LastName            string  `json:"lastName"`
	HouseholdMinistryID string  `json:"custentity_nscs_hh_ministry_id,omitempty"`
	Subsidiary          Related `json:"subsidiary"`
}

// PatchEmployeeRequest represents a employee update request
type PatchEmployeeRequest struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

// CreateEmployee handles POST /services/rest/record/v1/employee
func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req CreateEmployeeRequest
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
	if req.HouseholdMinistryID != "" {
		exists, err := h.store.ExternalIDExists(r.Context(), "employee", req.HouseholdMinistryID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, "Failed to check ministryId uniqueness")
			return
		}
		if exists {
			respondError(w, http.StatusConflict, domain.ErrorCodeDuplicateExternalID,
				"A record with the specified ministryId already exists")
			return
		}
	}

	// Create employee
	employee := models.NewEmployee()
	employee.FirstName = req.FirstName
	employee.LastName = req.LastName
	employee.HouseholdMinistryID = req.HouseholdMinistryID
	employee.Subsidiary = req.Subsidiary.ID

	created, err := h.store.CreateEmployee(r.Context(), employee)
	if err != nil {
		respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		return
	}

	log.Println("employee created, id=", created.ID)
	respondNoContent(w, r, strconv.FormatInt(created.ID, 10))
}

// GetEmployee handles GET /services/rest/record/v1/employee/{id}
func (h *Handler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Employee ID is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Invalid employee ID")
		return
	}

	employee, err := h.store.GetEmployee(r.Context(), id)
	if err != nil {
		if err.Error() == domain.ErrorCodeRecordNotFound {
			respondError(w, http.StatusNotFound, domain.ErrorCodeRecordNotFound, "Employee not found")
		} else {
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	log.Println("employee found, id=", employee.ID)
	respondSuccess(w, http.StatusOK, employee)
}
