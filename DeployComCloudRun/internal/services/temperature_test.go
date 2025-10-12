package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemperatureService_ConvertTemperatures(t *testing.T) {
	service := NewTemperatureService()

	tests := []struct {
		name      string
		celsius   float64
		expectedF float64
		expectedK float64
	}{
		{
			name:      "Zero celsius",
			celsius:   0,
			expectedF: 32,
			expectedK: 273,
		},
		{
			name:      "Positive temperature",
			celsius:   25,
			expectedF: 77,
			expectedK: 298,
		},
		{
			name:      "Negative temperature",
			celsius:   -10,
			expectedF: 14,
			expectedK: 263,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ConvertTemperatures(tt.celsius)

			assert.Equal(t, tt.celsius, result.TempC)
			assert.Equal(t, tt.expectedF, result.TempF)
			assert.Equal(t, tt.expectedK, result.TempK)
		})
	}
}
