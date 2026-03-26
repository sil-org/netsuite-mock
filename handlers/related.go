package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Related represents a Related Record instance that when serialized, outputs the id, refName, and links. When
// parsed, it expects either an internal ID or an external ID or simply a string or number representing the internal ID.
type Related struct {
	ID         int64
	ExternalID string
	refName    string
	href       string
}

// MarshalJSON serializes the Related struct into the expected JSON format for related records in NetSuite, including
// links and refName. It does not include the external ID in the output.
func (r Related) MarshalJSON() ([]byte, error) {
	type Link struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	}
	s := struct {
		Links   []Link `json:"links"`
		ID      string `json:"id"`
		RefName string `json:"refName"`
	}{
		Links: []Link{
			{
				Rel:  "self",
				Href: r.href,
			},
		},
		ID:      strconv.FormatInt(r.ID, 10),
		RefName: r.refName,
	}

	return json.Marshal(s)
}

// UnmarshalJSON parses the input JSON id or externalID. The refName and href values must be added after the fact if
// needed for serializing back to JSON.
func (r *Related) UnmarshalJSON(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch v := raw.(type) {
	case map[string]any:
		if idVal, ok := v["id"]; ok {
			i, err := parseID(idVal)
			if err != nil {
				return err
			}
			r.ID = i
		} else if idVal, ok = v["externalId"]; ok {
			r.ExternalID = idVal.(string)
		} else {
			return fmt.Errorf("invalid Related Record format: expected 'id' or 'externalId' field")
		}
	default:
		i, err := parseID(v)
		if err != nil {
			return err
		}
		r.ID = i
	}
	return nil
}

func parseID(input any) (int64, error) {
	switch v := input.(type) {
	case float64:
		if v == float64(int64(v)) {
			return int64(v), nil
		}
		return 0, fmt.Errorf("invalid ID: not a whole number")
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid ID: %w", err)
		}
		return id, nil
	default:
		return 0, fmt.Errorf("invalid ID type: %T", input)
	}
}
