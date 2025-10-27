package contains_letters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsLetters(t *testing.T) {
	cases := []struct {
		input    string
		expected bool
	}{
		{"abc", true},
		{"123", false},
		{"a1b2c3", true},
		{"", false},
		{"!@#$%^&*()", false},
		{"Привет", true},
		{"123你好", true},
		{"123456", false},
		{"   ", false},
	}

	for _, c := range cases {
		result := ContainsLetters(c.input)
		assert.Equal(t, result, c.expected)
	}
}
