package lib_test

import (
	"reflect"
	"testing"

	"github.com/netr/haki/lib"
)

func TestSplitQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single word",
			input:    "test",
			expected: []string{"test"},
		},
		{
			name:     "comma separated",
			input:    "netr,test,this",
			expected: []string{"netr", "test", "this"},
		},
		{
			name:     "comma separated with spaces",
			input:    "netr, test,    this",
			expected: []string{"netr", "test", "this"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "just spaces",
			input:    "   ",
			expected: []string{},
		},
		{
			name:     "trailing comma",
			input:    "test,",
			expected: []string{"test"},
		},
		{
			name:     "leading comma",
			input:    ",test",
			expected: []string{"test"},
		},
		{
			name:     "multiple commas",
			input:    "test,,hello",
			expected: []string{"test", "hello"},
		},
		{
			name:     "spaces between words without comma",
			input:    "hello world test",
			expected: []string{"hello world test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lib.SplitQuery(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("SplitQuery() = %v, want %v", got, tt.expected)
			}
		})
	}
}
