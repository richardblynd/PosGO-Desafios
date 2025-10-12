package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationService_ValidateZipcode(t *testing.T) {
	service := NewValidationService()

	tests := []struct {
		name     string
		zipcode  string
		expected bool
	}{
		{
			name:     "Valid zipcode",
			zipcode:  "01001000",
			expected: true,
		},
		{
			name:     "Valid zipcode with leading zero",
			zipcode:  "01310100",
			expected: true,
		},
		{
			name:     "Invalid zipcode - too short",
			zipcode:  "0100100",
			expected: false,
		},
		{
			name:     "Invalid zipcode - too long",
			zipcode:  "010010001",
			expected: false,
		},
		{
			name:     "Invalid zipcode - contains letters",
			zipcode:  "0100100A",
			expected: false,
		},
		{
			name:     "Invalid zipcode - contains hyphen",
			zipcode:  "01001-000",
			expected: false,
		},
		{
			name:     "Empty zipcode",
			zipcode:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ValidateZipcode(tt.zipcode)
			assert.Equal(t, tt.expected, result)
		})
	}
}
