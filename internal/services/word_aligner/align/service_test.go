package align

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlign(t *testing.T) {
	tests := []struct {
		name           string
		input          *Input
		mockHandler    func(w http.ResponseWriter, r *http.Request)
		expectedOutput *Output
		expectedError  string
	}{
		{
			name: "Success",
			input: &Input{
				SourceText: "Hello world",
				TargetText: "Hola mundo",
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/align", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&Output{
					SourceText: "Hello world",
					TargetText: "Hola mundo",
					Alignments: []Alignments{
						{SourceWord: "Hello", TargetWord: "Hola", SourceIndexes: [2]int{0, 5}, TargetIndexes: [2]int{0, 4}},
						{SourceWord: "world", TargetWord: "mundo", SourceIndexes: [2]int{6, 11}, TargetIndexes: [2]int{5, 10}},
					},
				})
			},
			expectedOutput: &Output{
				SourceText: "Hello world",
				TargetText: "Hola mundo",
				Alignments: []Alignments{
					{SourceWord: "Hello", TargetWord: "Hola", SourceIndexes: [2]int{0, 5}, TargetIndexes: [2]int{0, 4}},
					{SourceWord: "world", TargetWord: "mundo", SourceIndexes: [2]int{6, 11}, TargetIndexes: [2]int{5, 10}},
				},
			},
			expectedError: "",
		},
		{
			name: "Server Error",
			input: &Input{
				SourceText: "Hello world",
				TargetText: "Hola mundo",
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			},
			expectedOutput: nil,
			expectedError:  "status code: 500",
		},
		{
			name: "Invalid JSON Response",
			input: &Input{
				SourceText: "Hello world",
				TargetText: "Hola mundo",
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			},
			expectedOutput: nil,
			expectedError:  "invalid character 'i' looking for beginning of value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.mockHandler))
			defer server.Close()

			service := New(server.URL)

			// We need to override the URL in the service instance because New() sets it to the input URL
			// which is what we want.
			// However, in the loop, we are creating a new server for each test case.
			// So `service.URL` should be `server.URL`.
			// The `New` function sets `service.URL`.

			output, err := service.Align(tt.input)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}
