package get_chapter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetChapter(t *testing.T) {
	tests := []struct {
		name           string
		inputPath      string
		mockS3         func(*MockS3)
		expectedOutput Output
		expectedError  error
	}{
		{
			name:      "Success",
			inputPath: "chapter/1",
			mockS3: func(m *MockS3) {
				m.On("Get", "chapter/1").Return([]byte("chapter content"), nil)
			},
			expectedOutput: []byte("chapter content"),
			expectedError:  nil,
		},
		{
			name:      "S3 Error",
			inputPath: "chapter/error",
			mockS3: func(m *MockS3) {
				m.On("Get", "chapter/error").Return(nil, errors.New("s3 error"))
			},
			expectedOutput: nil,
			expectedError:  errors.New("failed to get chapter"),
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
			output, err := service.GetChapter(input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}
