package translate

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTranslate(t *testing.T) {
	tests := []struct {
		name           string
		inputQ         string
		inputSource    domain.SupportedLang
		inputTarget    domain.SupportedLang
		mockClient     func(*MockHTTPClient)
		expectedOutput *Output
		expectedError  string
	}{
		{
			name:        "Success",
			inputQ:      "hello",
			inputSource: "en",
			inputTarget: "es",
			mockClient: func(m *MockHTTPClient) {
				responseBody := `{"translatedText": "hola"}`
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
				}
				m.On("PostForm", "http://localhost:5000/translate", mock.Anything).Return(resp, nil)
			},
			expectedOutput: &Output{TranslatedText: "hola"},
			expectedError:  "",
		},
		{
			name:        "API Error",
			inputQ:      "hello",
			inputSource: "en",
			inputTarget: "es",
			mockClient: func(m *MockHTTPClient) {
				m.On("PostForm", "http://localhost:5000/translate", mock.Anything).Return(nil, errors.New("connection refused"))
			},
			expectedOutput: nil,
			expectedError:  "translate.Service.Translate: All attempts fail:\n#1: connection refused\n#2: connection refused\n#3: connection refused\n#4: connection refused\n#5: connection refused",
		},
		{
			name:        "Non-200 Status Code",
			inputQ:      "hello",
			inputSource: "en",
			inputTarget: "es",
			mockClient: func(m *MockHTTPClient) {
				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}
				m.On("PostForm", "http://localhost:5000/translate", mock.Anything).Return(resp, nil)
			},
			expectedOutput: nil,
			expectedError:  "translate.Service.Translate: status code: 500",
		},
		{
			name:        "JSON Unmarshal Error",
			inputQ:      "hello",
			inputSource: "en",
			inputTarget: "es",
			mockClient: func(m *MockHTTPClient) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("invalid json")),
				}
				m.On("PostForm", "http://localhost:5000/translate", mock.Anything).Return(resp, nil)
			},
			expectedOutput: nil,
			expectedError:  "translate.Service.Translate: invalid character 'i' looking for beginning of value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockHTTPClient)
			if tt.mockClient != nil {
				tt.mockClient(mockClient)
			}

			service := New("http://localhost:5000", mockClient)

			input := &Input{
				Q:      tt.inputQ,
				Source: tt.inputSource,
				Target: tt.inputTarget,
			}

			output, err := service.Translate(input)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, tt.expectedOutput.TranslatedText, output.TranslatedText)
			}
		})
	}
}
