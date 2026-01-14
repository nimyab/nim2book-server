package models

import (
	"testing"
)

func TestStringArray_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    StringArray
		expected string
	}{
		{
			name:     "nil array",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty array",
			input:    StringArray{},
			expected: "{}",
		},
		{
			name:     "single element",
			input:    StringArray{"chapter1.json"},
			expected: `{"chapter1.json"}`,
		},
		{
			name:     "multiple elements",
			input:    StringArray{"chapter1.json", "chapter2.json", "chapter3.json"},
			expected: `{"chapter1.json","chapter2.json","chapter3.json"}`,
		},
		{
			name:     "element with quotes",
			input:    StringArray{`chapter"1".json`},
			expected: `{"chapter\"1\".json"}`,
		},
		{
			name:     "element with backslash",
			input:    StringArray{`chapter\1.json`},
			expected: `{"chapter\\1.json"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.input.Value()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.input == nil {
				if value != nil {
					t.Errorf("expected nil, got %v", value)
				}
				return
			}

			strValue, ok := value.(string)
			if !ok {
				t.Fatalf("expected string value, got %T", value)
			}

			if strValue != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, strValue)
			}
		})
	}
}

func TestStringArray_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected StringArray
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty array string",
			input:    "{}",
			expected: StringArray{},
		},
		{
			name:     "empty array bytes",
			input:    []byte("{}"),
			expected: StringArray{},
		},
		{
			name:     "single element string",
			input:    `{"chapter1.json"}`,
			expected: StringArray{"chapter1.json"},
		},
		{
			name:     "single element bytes",
			input:    []byte(`{"chapter1.json"}`),
			expected: StringArray{"chapter1.json"},
		},
		{
			name:     "multiple elements",
			input:    `{"chapter1.json","chapter2.json","chapter3.json"}`,
			expected: StringArray{"chapter1.json", "chapter2.json", "chapter3.json"},
		},
		{
			name:     "elements with quotes",
			input:    `{"chapter\"1\".json"}`,
			expected: StringArray{`chapter"1".json`},
		},
		{
			name:     "elements with backslash",
			input:    `{"chapter\\1.json"}`,
			expected: StringArray{`chapter\1.json`},
		},
		{
			name:     "unquoted elements (PostgreSQL format)",
			input:    `{chapter1.json,chapter2.json}`,
			expected: StringArray{"chapter1.json", "chapter2.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result StringArray
			err := result.Scan(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(result))
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("element %d: expected %q, got %q", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestStringArray_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input StringArray
	}{
		{
			name:  "empty array",
			input: StringArray{},
		},
		{
			name:  "single element",
			input: StringArray{"chapter1.json"},
		},
		{
			name:  "multiple elements",
			input: StringArray{"chapter1.json", "chapter2.json", "chapter3.json"},
		},
		{
			name:  "complex paths",
			input: StringArray{"path/to/chapter1.json", "another/path/chapter2.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to database value
			value, err := tt.input.Value()
			if err != nil {
				t.Fatalf("Value() error: %v", err)
			}

			// Convert back from database value
			var result StringArray
			err = result.Scan(value)
			if err != nil {
				t.Fatalf("Scan() error: %v", err)
			}

			// Compare
			if len(result) != len(tt.input) {
				t.Fatalf("expected length %d, got %d", len(tt.input), len(result))
			}

			for i := range result {
				if result[i] != tt.input[i] {
					t.Errorf("element %d: expected %q, got %q", i, tt.input[i], result[i])
				}
			}
		})
	}
}

func TestBook_GetChapterCount(t *testing.T) {
	tests := []struct {
		name     string
		book     Book
		expected int
	}{
		{
			name: "empty chapters",
			book: Book{
				ChapterPaths: StringArray{},
			},
			expected: 0,
		},
		{
			name: "single chapter",
			book: Book{
				ChapterPaths: StringArray{"chapter1.json"},
			},
			expected: 1,
		},
		{
			name: "multiple chapters",
			book: Book{
				ChapterPaths: StringArray{"ch1.json", "ch2.json", "ch3.json"},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := tt.book.GetChapterCount()
			if count != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, count)
			}
		})
	}
}

func TestBook_HasCover(t *testing.T) {
	emptyCover := ""
	validCover := "cover.jpg"

	tests := []struct {
		name     string
		book     Book
		expected bool
	}{
		{
			name: "nil cover",
			book: Book{
				Cover: nil,
			},
			expected: false,
		},
		{
			name: "empty cover",
			book: Book{
				Cover: &emptyCover,
			},
			expected: false,
		},
		{
			name: "valid cover",
			book: Book{
				Cover: &validCover,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasCover := tt.book.HasCover()
			if hasCover != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, hasCover)
			}
		})
	}
}
