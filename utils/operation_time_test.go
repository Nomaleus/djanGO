package utils

import (
	"testing"
)

func TestGetOperationTime(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		minTime   int
		maxTime   int
	}{
		{
			name:      "Addition",
			operation: "+",
			minTime:   100,
			maxTime:   2000,
		},
		{
			name:      "Multiplication",
			operation: "*",
			minTime:   100,
			maxTime:   4000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			time := GetOperationTime(tt.operation)
			if time < tt.minTime {
				t.Errorf("Time %d is less than minimum %d", time, tt.minTime)
			}
			if time > tt.maxTime {
				t.Errorf("Time %d is greater than maximum %d", time, tt.maxTime)
			}
		})
	}
}
