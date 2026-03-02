package connector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPageTokenToInt(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		want      int
		wantError string
	}{
		{
			name:  "empty token returns 0",
			token: "",
			want:  0,
		},
		{
			name:  "valid page number",
			token: "5",
			want:  5,
		},
		{
			name:  "page 1",
			token: "1",
			want:  1,
		},
		{
			name:  "page 0",
			token: "0",
			want:  0,
		},
		{
			name:  "large page number",
			token: "9999",
			want:  9999,
		},
		{
			name:  "negative number",
			token: "-1",
			want:  -1,
		},
		{
			name:      "invalid non-numeric token",
			token:     "abc",
			want:      0,
			wantError: "baton-buildkite: invalid page token: strconv.Atoi: parsing \"abc\": invalid syntax",
		},
		{
			name:      "invalid mixed token",
			token:     "12abc",
			want:      0,
			wantError: "baton-buildkite: invalid page token: strconv.Atoi: parsing \"12abc\": invalid syntax",
		},
		{
			name:      "invalid float",
			token:     "1.5",
			want:      0,
			wantError: "baton-buildkite: invalid page token: strconv.Atoi: parsing \"1.5\": invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pageTokenToInt(tt.token)
			if tt.wantError != "" {
				assert.EqualError(t, err, tt.wantError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
