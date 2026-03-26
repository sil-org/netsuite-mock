package models

// Error represents a NetSuite API error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// ErrorResponse wraps an error in the API response format
type ErrorResponse struct {
	Error *Error `json:"error"`
}

// NewError creates a new error
func NewError(code, message, errorType string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Type:    errorType,
	}
}

// NewErrorResponse wraps an error in a response
func NewErrorResponse(code, message, errorType string) *ErrorResponse {
	return &ErrorResponse{
		Error: NewError(code, message, errorType),
	}
}

// DataResponse wraps data in the standard API response format
type DataResponse struct {
	Data any `json:"data"`
}

// QueryResponse wraps query results in the standard API response format
type QueryResponse struct {
	Count        int              `json:"count"`
	HasMore      bool             `json:"hasMore"`
	Items        []map[string]any `json:"items"`
	Offset       int              `json:"offset"`
	TotalResults int              `json:"totalResults"`
}

// NewQueryResponse creates a new query response
func NewQueryResponse() *QueryResponse {
	return &QueryResponse{
		Items: []map[string]any{},
	}
}
