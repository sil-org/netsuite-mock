package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/sil-org/netsuite-mock/domain"
	"github.com/sil-org/netsuite-mock/models"
	"github.com/sil-org/netsuite-mock/storage"
)

// QueryRequest represents a SuiteQL query request
type QueryRequest struct {
	Q string `json:"q"`
}

var errInvalidQuerySyntax = errors.New("INVALID_QUERY_SYNTAX")

var whereEqualsPattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.+)$`)

// QueryHandler handles POST /services/rest/query/v1/suiteql
func (h *Handler) QueryHandler(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Failed to parse request body")
		return
	}

	if req.Q == "" {
		respondError(w, http.StatusBadRequest, domain.ErrorCodeInvalidRequest, "Query parameter 'q' is required")
		return
	}

	// Parse the query
	results, err := h.executeQuery(r.Context(), req.Q)
	if err != nil {
		switch {
		case errors.Is(err, errInvalidQuerySyntax):
			respondError(w, http.StatusBadRequest, "INVALID_QUERY_SYNTAX", err.Error())
		case errors.Is(err, storage.ErrInvalidQueryFilter):
			respondError(w, http.StatusBadRequest, "INVALID_QUERY_FILTER", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, domain.ErrorCodeInternalError, err.Error())
		}
		return
	}

	respondQuery(w, http.StatusOK, results)
}

// executeQuery parses and executes a simple SuiteQL query
func (h *Handler) executeQuery(ctx context.Context, query string) ([]map[string]any, error) {
	// Simple parser for basic SELECT queries
	// Format: SELECT field1,field2 FROM recordType WHERE field = 'value'

	query = strings.TrimSpace(query)
	if !strings.HasPrefix(strings.ToUpper(query), "SELECT") {
		return nil, fmt.Errorf("query must start with SELECT: %w", errInvalidQuerySyntax)
	}

	// Parse SELECT clause
	selectIdx := strings.ToUpper(query)
	fromIdx := strings.Index(selectIdx, "FROM")
	if fromIdx == -1 {
		return nil, fmt.Errorf("missing FROM clause: %w", errInvalidQuerySyntax)
	}

	selectPart := strings.TrimSpace(query[6:fromIdx])
	selectFields := strings.Split(selectPart, ",")
	for i := range selectFields {
		selectFields[i] = strings.TrimSpace(selectFields[i])
	}

	// Parse FROM clause
	fromPart := strings.TrimSpace(query[fromIdx+4:])
	whereIdx := strings.Index(strings.ToUpper(fromPart), "WHERE")

	var recordType string
	var whereClause string

	if whereIdx == -1 {
		recordType = strings.TrimSpace(fromPart)
	} else {
		recordType = strings.TrimSpace(fromPart[:whereIdx])
		whereClause = strings.TrimSpace(fromPart[whereIdx+5:])
	}

	recordType = strings.ToLower(recordType)

	// Parse WHERE clause if present
	var filters []*storage.QueryFilter
	if whereClause != "" {
		match := whereEqualsPattern.FindStringSubmatch(whereClause)
		if len(match) != 3 {
			return nil, fmt.Errorf(
				"unsupported WHERE clause %q (only single '=' predicates are supported): %w",
				whereClause,
				errInvalidQuerySyntax,
			)
		}

		field := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2])
		if value == "" {
			return nil, fmt.Errorf("WHERE value is required: %w", errInvalidQuerySyntax)
		}

		// Remove single quotes if present
		if strings.HasPrefix(value, "'") || strings.HasSuffix(value, "'") {
			if !strings.HasPrefix(value, "'") || !strings.HasSuffix(value, "'") {
				return nil, fmt.Errorf("malformed quoted value in WHERE clause: %w", errInvalidQuerySyntax)
			}
			value = value[1 : len(value)-1]
		}

		filters = append(filters, &storage.QueryFilter{
			Field:    field,
			Operator: "=",
			Values:   []any{value},
		})
	}

	// Execute query based on record type
	switch recordType {
	case "customer":
		customers, err := h.store.QueryCustomers(ctx, filters)
		if err != nil {
			return nil, err
		}
		return customersToMaps(selectFields, customers), nil

	case "employee":
		employees, err := h.store.QueryEmployees(ctx, filters)
		if err != nil {
			return nil, err
		}
		return employeesToMaps(selectFields, employees), nil

	case "account":
		accounts, err := h.store.QueryAccounts(ctx, filters)
		if err != nil {
			return nil, err
		}
		return accountsToMaps(selectFields, accounts), nil

	case "customrecord_cseg3":
		// SuiteKey records - for now, return empty
		return []map[string]any{}, nil

	default:
		return nil, fmt.Errorf("unsupported record type %q: %w", recordType, errInvalidQuerySyntax)
	}
}

// customersToMaps converts customers to maps with only selected fields
func customersToMaps(fields []string, customers []*models.Customer) []map[string]any {
	results := make([]map[string]any, 0, len(customers))

	for _, customer := range customers {
		m := make(map[string]any)
		for _, field := range fields {
			switch strings.ToLower(field) {
			case "id":
				m["id"] = strconv.Itoa(int(customer.ID))
			case "externalid":
				m["externalid"] = customer.ExternalID
			case "firstname":
				m["firstname"] = customer.FirstName
			case "lastname":
				m["lastname"] = customer.LastName
			case "subsidiary":
				m["subsidiary"] = customer.Subsidiary
			case "isinactive":
				m["isinactive"] = boolToString(customer.IsInactive)
			}
		}
		results = append(results, m)
	}

	return results
}

// employeesToMaps converts employees to maps with only selected fields
func employeesToMaps(fields []string, employees []*models.Employee) []map[string]any {
	results := make([]map[string]any, 0, len(employees))

	for _, employee := range employees {
		m := make(map[string]any)
		for _, field := range fields {
			switch strings.ToLower(field) {
			case "id":
				m["id"] = strconv.Itoa(int(employee.ID))
			case "firstname":
				m["firstname"] = employee.FirstName
			case "lastname":
				m["lastname"] = employee.LastName
			case "custentity_nscs_hh_ministry_id":
				m["custentity_nscs_hh_ministry_id"] = employee.HouseholdMinistryID
			case "subsidiary":
				m["subsidiary"] = employee.Subsidiary
			}
		}
		results = append(results, m)
	}

	return results
}

// accountsToMaps converts accounts to maps with only selected fields
func accountsToMaps(fields []string, accounts []*models.Account) []map[string]any {
	results := make([]map[string]any, 0, len(accounts))

	for _, account := range accounts {
		m := make(map[string]any)
		for _, field := range fields {
			switch strings.ToLower(field) {
			case "id":
				m["id"] = strconv.Itoa(int(account.ID))
			case "externalid":
				m["externalid"] = account.ExternalID
			case "acctnumber":
				m["acctnumber"] = account.AccountNumber
			case "class":
				m["class"] = account.Class
			}
		}
		results = append(results, m)
	}

	return results
}

// boolToString converts a bool to a SuiteQL-specific string, specifically "T" or "F"
func boolToString(b bool) string {
	if b {
		return "T"
	}
	return "F"
}
