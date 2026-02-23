package file_public

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFile(t *testing.T) {
	tests := []struct {
		name           string
		inputPath      string
		mockS3         func(*MockS3)
		expectedOutput Output
		expectedError  string
	}{
		{
			name:      "Success",
			inputPath: "path/to/file",
			mockS3: func(m *MockS3) {
				m.On("Get", "path/to/file").Return([]byte("file content"), nil)
			},
			expectedOutput: Output([]byte("file content")),
			expectedError:  "",
		},
		{
			name:      "Error",
			inputPath: "path/to/file",
			mockS3: func(m *MockS3) {
				m.On("Get", "path/to/file").Return(nil, errors.New("s3 error"))
			},
			expectedOutput: nil,
			expectedError:  "failed to get file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := NewMockS3(t)
			if tt.mockS3 != nil {
				tt.mockS3(mockS3)
			}

			service := New(mockS3)

			input := &Input{Path: tt.inputPath}
			output, err := service.GetFile(input)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}
