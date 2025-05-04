package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsOrderNumberValid(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "test valid number",
			number: "1488342146863",
			want:   true,
		},
		{
			name:   "test invalid number",
			number: "123",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOrderNumberValid(tt.number)
			assert.Equal(t, tt.want, got, "IsOrderNumberValid() = %v, want %v", got, tt.want)
		})
	}
}
