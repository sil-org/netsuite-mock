package handlers

import (
	"encoding/json"
	"testing"
)

func TestRelatedUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Related
		wantErr bool
	}{
		{
			name:    "whole number",
			input:   "100",
			want:    Related{ID: 100},
			wantErr: false,
		},
		{
			name:    "number with zero fractional part",
			input:   "101.0",
			want:    Related{ID: 101},
			wantErr: false,
		},
		{
			name:    "number with non-zero fractional part",
			input:   "100.5",
			want:    Related{ID: 0},
			wantErr: true,
		},
		{
			name:    "string",
			input:   "102",
			want:    Related{ID: 102},
			wantErr: false,
		},
		{
			name:    "string invalid",
			input:   "100.5",
			want:    Related{ID: 0},
			wantErr: true,
		},
		{
			name:    "object with numeric string id",
			input:   `{"id": "103"}`,
			want:    Related{ID: 103},
			wantErr: false,
		},
		{
			name:    "object with numeric id",
			input:   `{"id": 104}`,
			want:    Related{ID: 104},
			wantErr: false,
		},
		{
			name:    "object with externalId",
			input:   `{"externalId": "EID8000"}`,
			want:    Related{ExternalID: "EID8000"},
			wantErr: false,
		},
		{
			name:    "object with id and externalId",
			input:   `{"id":105, "externalId": "EID9000"}`,
			want:    Related{ID: 105},
			wantErr: false,
		},
		{
			name:    "object without id",
			input:   `{}`,
			want:    Related{ID: 0},
			wantErr: true,
		},
		{
			name:    "null",
			input:   "null",
			want:    Related{ID: 0},
			wantErr: true,
		},
		{
			name:    "boolean",
			input:   "true",
			want:    Related{ID: 0},
			wantErr: true,
		},
		{
			name:    "array",
			input:   "[100]",
			want:    Related{ID: 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Related
			err := json.Unmarshal([]byte(tt.input), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelatedMarshalJSON(t *testing.T) {
	id := Related{
		ID:         100,
		ExternalID: "EID1234",
		refName:    "example-refName",
		href:       "https://example.com/related/100",
	}
	data, err := id.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
	}

	const want = `{"links":[{"rel":"self","href":"https://example.com/related/100"}],"id":"100","refName":"example-refName"}`
	if string(data) != want {
		t.Errorf("MarshalJSON() = %s, want %s", data, want)
	}
}
