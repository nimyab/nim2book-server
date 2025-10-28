package translate_book

import (
	"errors"
	"testing"
	"time"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/services/libretranslate/translate"
	"github.com/nimyab/nim2book-back/internal/services/word_aligner/align"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_translateAndAlignParagraph(t *testing.T) {
	tests := []struct {
		name           string
		paragraph      string
		from           domain.SupportedLang
		to             domain.SupportedLang
		translatorResp *translate.Output
		translatorErr  error
		alignerResp    *align.Output
		alignerErr     error
		wantErr        bool
	}{
		{
			name:      "successful translation and alignment",
			paragraph: "Hello world",
			from:      domain.En,
			to:        domain.Ru,
			translatorResp: &translate.Output{
				TranslatedText: "Привет мир",
			},
			translatorErr: nil,
			alignerResp: &align.Output{
				Alignments: []align.Alignments{
					{
						SourceWord:    "Hello",
						TargetWord:    "Привет",
						SourceIndexes: [2]int{0, 4},
						TargetIndexes: [2]int{0, 5},
					},
					{
						SourceWord:    "world",
						TargetWord:    "мир",
						SourceIndexes: [2]int{6, 10},
						TargetIndexes: [2]int{7, 9},
					},
				},
			},
			alignerErr: nil,
			wantErr:    false,
		},
		{
			name:           "translation error",
			paragraph:      "Hello world",
			from:           domain.En,
			to:             domain.Ru,
			translatorResp: nil,
			translatorErr:  errors.New("translation failed"),
			wantErr:        true,
		},
		{
			name:      "alignment error",
			paragraph: "Hello world",
			from:      domain.En,
			to:        domain.Ru,
			translatorResp: &translate.Output{
				TranslatedText: "Привет мир",
			},
			translatorErr: nil,
			alignerResp:   nil,
			alignerErr:    errors.New("alignment failed"),
			wantErr:       true,
		},
		{
			name:      "paragraph without letters",
			paragraph: "123 456",
			from:      domain.En,
			to:        domain.Ru,
			translatorResp: &translate.Output{
				TranslatedText: "123 456",
			},
			translatorErr: nil,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := new(MockS3)
			mockPg := new(MockPostgres)
			mockWordAligner := new(MockWordAligner)
			mockTranslator := new(MockTranslator)
			mockNotificationSender := new(MockNotificationSender)

			s := New(
				mockS3,
				mockPg,
				mockWordAligner,
				mockTranslator,
				5,
				10*time.Millisecond,
				mockNotificationSender,
			)

			// Setup expectations
			mockTranslator.On("Translate", mock.MatchedBy(func(input *translate.Input) bool {
				return input.Q == tt.paragraph && input.Source == tt.from && input.Target == tt.to
			})).Return(tt.translatorResp, tt.translatorErr)

			if tt.translatorErr == nil && tt.paragraph != "123 456" {
				mockWordAligner.On("Align", mock.MatchedBy(func(input *align.Input) bool {
					return input.SourceText == tt.paragraph
				})).Return(tt.alignerResp, tt.alignerErr)
			}

			result, err := s.translateAndAlignParagraph(tt.paragraph, tt.from, tt.to)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.paragraph, result.OriginalParagraph)
				if tt.translatorResp != nil {
					assert.Equal(t, tt.translatorResp.TranslatedText, result.TranslatedParagraph)
				}
			}

			mockTranslator.AssertExpectations(t)
			if tt.translatorErr == nil && tt.paragraph != "123 456" {
				mockWordAligner.AssertExpectations(t)
			}
		})
	}
}

func TestService_saveChapterToS3(t *testing.T) {
	tests := []struct {
		name      string
		chapter   *domain.ChapterAlignNode
		bookTitle string
		s3Err     error
		wantErr   bool
	}{
		{
			name: "successful save",
			chapter: &domain.ChapterAlignNode{
				Id:              "chapter1",
				Order:           0,
				Title:           "Chapter 1",
				TranslatedTitle: "Глава 1",
				Content: []domain.ParagraphAlignNode{
					{
						OriginalParagraph:   "Hello",
						TranslatedParagraph: "Привет",
					},
				},
			},
			bookTitle: "Test Book",
			s3Err:     nil,
			wantErr:   false,
		},
		{
			name: "s3 upload error",
			chapter: &domain.ChapterAlignNode{
				Id:    "chapter1",
				Order: 0,
			},
			bookTitle: "Test Book",
			s3Err:     errors.New("upload failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := new(MockS3)
			mockPg := new(MockPostgres)
			mockWordAligner := new(MockWordAligner)
			mockTranslator := new(MockTranslator)
			mockNotificationSender := new(MockNotificationSender)

			s := New(
				mockS3,
				mockPg,
				mockWordAligner,
				mockTranslator,
				5,
				10*time.Millisecond,
				mockNotificationSender,
			)

			mockS3.On("Upload", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(tt.s3Err)

			path, err := s.saveChapterToS3(tt.chapter, tt.bookTitle)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, path)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, path)
				assert.Contains(t, path, "Test_Book")
			}

			mockS3.AssertExpectations(t)
		})
	}
}

func TestService_saveCoverToS3(t *testing.T) {
	tests := []struct {
		name      string
		coverData []byte
		bookTitle string
		s3Err     error
		wantErr   bool
	}{
		{
			name:      "successful save",
			coverData: []byte("cover image data"),
			bookTitle: "Test Book",
			s3Err:     nil,
			wantErr:   false,
		},
		{
			name:      "s3 upload error",
			coverData: []byte("cover image data"),
			bookTitle: "Test Book",
			s3Err:     errors.New("upload failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := new(MockS3)
			mockPg := new(MockPostgres)
			mockWordAligner := new(MockWordAligner)
			mockTranslator := new(MockTranslator)
			mockNotificationSender := new(MockNotificationSender)

			s := New(
				mockS3,
				mockPg,
				mockWordAligner,
				mockTranslator,
				5,
				10*time.Millisecond,
				mockNotificationSender,
			)

			mockS3.On("Upload", mock.AnythingOfType("string"), tt.coverData).Return(tt.s3Err)

			path, err := s.saveCoverToS3(tt.coverData, tt.bookTitle)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, path)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, path)
				assert.Contains(t, path, "cover/Test_Book")
			}

			mockS3.AssertExpectations(t)
		})
	}
}

func TestService_checkChapterInStorage(t *testing.T) {
	tests := []struct {
		name         string
		chapterOrder int
		bookTitle    string
		s3Err        error
		wantPath     string
	}{
		{
			name:         "chapter exists",
			chapterOrder: 0,
			bookTitle:    "Test Book",
			s3Err:        nil,
			wantPath:     "book/Test_Book/0.json",
		},
		{
			name:         "chapter does not exist",
			chapterOrder: 0,
			bookTitle:    "Test Book",
			s3Err:        errors.New("not found"),
			wantPath:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := new(MockS3)
			mockPg := new(MockPostgres)
			mockWordAligner := new(MockWordAligner)
			mockTranslator := new(MockTranslator)
			mockNotificationSender := new(MockNotificationSender)

			s := New(
				mockS3,
				mockPg,
				mockWordAligner,
				mockTranslator,
				5,
				10*time.Millisecond,
				mockNotificationSender,
			)

			expectedPath := "book/Test_Book/0.json"
			mockS3.On("Check", expectedPath).Return(tt.s3Err)

			path := s.checkChapterInStorage(tt.chapterOrder, tt.bookTitle)

			assert.Equal(t, tt.wantPath, path)
			mockS3.AssertExpectations(t)
		})
	}
}

func TestNew(t *testing.T) {
	mockS3 := new(MockS3)
	mockPg := new(MockPostgres)
	mockWordAligner := new(MockWordAligner)
	mockTranslator := new(MockTranslator)
	mockNotificationSender := new(MockNotificationSender)

	maxRequestCount := 10
	waitDuration := 100 * time.Millisecond

	s := New(
		mockS3,
		mockPg,
		mockWordAligner,
		mockTranslator,
		maxRequestCount,
		waitDuration,
		mockNotificationSender,
	)

	assert.NotNil(t, s)
	assert.Equal(t, maxRequestCount, s.maxRequestCount)
	assert.Equal(t, waitDuration, s.waitDuration)
	assert.Equal(t, 0, s.currentCountBookTranslating)
	assert.Equal(t, mockS3, s.s3)
	assert.Equal(t, mockPg, s.pg)
	assert.Equal(t, mockWordAligner, s.wordAligner)
	assert.Equal(t, mockTranslator, s.translator)
	assert.Equal(t, mockNotificationSender, s.notificationSignal)
}
