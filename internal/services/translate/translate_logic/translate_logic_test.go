package translated_logic

import (
	"context"
	"errors"
	"testing"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/services/libretranslate/translate"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
	pb "github.com/nimyab/nim2book-back/proto/word_aligner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/timsims/pamphlet"
)

func TestTranslateAndAlignParagraph(t *testing.T) {
	tests := []struct {
		name              string
		paragraph         string
		mockTranslator    func(*MockTranslator)
		mockWordAligner   func(*MockWordAligner)
		mockThrottler     func(*MockThrottler)
		expectedParagraph domain.ParagraphAlignNode
		expectedError     error
	}{
		{
			name:      "Success",
			paragraph: "Hello world",
			mockTranslator: func(m *MockTranslator) {
				m.On("Translate", &translate.Input{
					Q:      "Hello world",
					Source: domain.En,
					Target: domain.Ru,
				}).Return(&translate.Output{TranslatedText: "Привет мир"}, nil)
			},
			mockWordAligner: func(m *MockWordAligner) {
				m.On("Align", mock.Anything, &pb.AlignRequest{
					SourceText: "Hello world",
					TargetText: "Привет мир",
				}).Return(&pb.AlignResponse{
					Alignments: []*pb.AlignmentResult{
						{SrcStart: 0, SrcEnd: 5, TargetStart: 0, TargetEnd: 6},
					},
				}, nil)
			},
			mockThrottler: func(m *MockThrottler) {
				m.On("Throttle").Times(2)
			},
			expectedParagraph: domain.ParagraphAlignNode{
				OriginalParagraph:   "Hello world",
				TranslatedParagraph: "Привет мир",
				AlignmentWords: []domain.WordAlignNode{
					{IndexesOriginalWord: [2]int{0, 5}, IndexesTranslatedWord: [2]int{0, 6}},
				},
			},
			expectedError: nil,
		},
		{
			name:      "No Letters in Original",
			paragraph: "...",
			mockTranslator: func(m *MockTranslator) {
				m.On("Translate", &translate.Input{
					Q:      "...",
					Source: domain.En,
					Target: domain.Ru,
				}).Return(&translate.Output{TranslatedText: "..."}, nil)
			},
			mockWordAligner: nil,
			mockThrottler:   nil,
			expectedParagraph: domain.ParagraphAlignNode{
				OriginalParagraph:   "...",
				TranslatedParagraph: "...",
				AlignmentWords: []domain.WordAlignNode{
					{IndexesOriginalWord: [2]int{0, 2}, IndexesTranslatedWord: [2]int{0, 2}},
				},
			},
			expectedError: nil,
		},
		{
			name:      "Translation Error",
			paragraph: "Hello",
			mockTranslator: func(m *MockTranslator) {
				m.On("Translate", mock.Anything).Return(nil, errors.New("translate error"))
			},
			mockWordAligner: nil,
			mockThrottler:   nil,
			expectedError:   errors.New("failed to translate paragraph"),
		},
		{
			name:      "Alignment Error",
			paragraph: "Hello",
			mockTranslator: func(m *MockTranslator) {
				m.On("Translate", mock.Anything).Return(&translate.Output{TranslatedText: "Привет"}, nil)
			},
			mockWordAligner: func(m *MockWordAligner) {
				m.On("Align", mock.Anything, mock.Anything).Return(nil, errors.New("align error"))
			},
			mockThrottler: func(m *MockThrottler) {
				m.On("Throttle").Times(2)
			},
			expectedError: errors.New("failed to align words"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewMockTranslator(t)
			wordAligner := NewMockWordAligner(t)
			throttler := NewMockThrottler(t)

			if tt.mockTranslator != nil {
				tt.mockTranslator(translator)
			}
			if tt.mockWordAligner != nil {
				tt.mockWordAligner(wordAligner)
			}
			if tt.mockThrottler != nil {
				tt.mockThrottler(throttler)
			}

			l := New(translator, wordAligner)
			res, err := l.TranslateAndAlignParagraph(context.Background(), tt.paragraph, domain.En, domain.Ru, throttler)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedParagraph, res)
			}
		})
	}
}

func TestTranslateChapter(t *testing.T) {
	tests := []struct {
		name            string
		chapter         epub_parser.FormattedChapter
		mockTranslator  func(*MockTranslator)
		mockWordAligner func(*MockWordAligner)
		mockThrottler   func(*MockThrottler)
		expectedResult  *domain.ChapterAlignNode
		expectedError   string
	}{
		{
			name: "Success",
			chapter: epub_parser.FormattedChapter{
				Chapter: pamphlet.Chapter{
					ID:    "1",
					Title: "Chapter 1",
				},
				Content: []epub_parser.ContentItem{
					{
						Type:     epub_parser.ContentTypeText,
						TextNode: &epub_parser.TextNode{Text: "Hello"},
					},
				},
			},
			mockTranslator: func(m *MockTranslator) {
				// Translate Paragraph
				m.On("Translate", &translate.Input{
					Q:      "Hello",
					Source: domain.En,
					Target: domain.Ru,
				}).Return(&translate.Output{TranslatedText: "Привет"}, nil)

				// Translate Title
				m.On("Translate", &translate.Input{
					Q:      "Chapter 1",
					Source: domain.En,
					Target: domain.Ru,
				}).Return(&translate.Output{TranslatedText: "Глава 1"}, nil)
			},
			mockWordAligner: func(m *MockWordAligner) {
				m.On("Align", mock.Anything, &pb.AlignRequest{
					SourceText: "Hello",
					TargetText: "Привет",
				}).Return(&pb.AlignResponse{
					Alignments: []*pb.AlignmentResult{
						{SrcStart: 0, SrcEnd: 5, TargetStart: 0, TargetEnd: 6},
					},
				}, nil)
			},
			mockThrottler: func(m *MockThrottler) {
				m.On("Throttle").Times(2)
			},
			expectedResult: &domain.ChapterAlignNode{
				Id:              "1",
				Title:           "Chapter 1",
				TranslatedTitle: "Глава 1",
				Content: []domain.ContentNode{
					{
						Type: domain.ParagraphAlignNodeTypeParagraph,
						ParagraphAlignNode: &domain.ParagraphAlignNode{
							OriginalParagraph:   "Hello",
							TranslatedParagraph: "Привет",
							AlignmentWords: []domain.WordAlignNode{
								{IndexesOriginalWord: [2]int{0, 5}, IndexesTranslatedWord: [2]int{0, 6}},
							},
						},
					},
				},
			},
			expectedError: "",
		},
		/*
			{
				name: "With Image",
				chapter: epub_parser.FormattedChapter{
					Chapter: pamphlet.Chapter{
						ID:    "img1",
						Title: "Image Chapter",
					},
					Content: []epub_parser.ContentItem{
						{
							Type:     epub_parser.ContentTypeText,
							TextNode: &epub_parser.TextNode{Text: "Text"},
						},
						{
							Type: epub_parser.ContentTypeImage,
							ImageNode: &epub_parser.ImageNode{
								// ImageData: []byte("fake"),
							},
						},
					},
				},
				mockTranslator: func(m *MockTranslator) {
					m.On("Translate", &translate.Input{
						Q:      "Text",
						Source: domain.En,
						Target: domain.Ru,
					}).Return(&translate.Output{TranslatedText: "Текст"}, nil)

					m.On("Translate", &translate.Input{
						Q:      "Image Chapter",
						Source: domain.En,
						Target: domain.Ru,
					}).Return(&translate.Output{TranslatedText: "Глава с картинкой"}, nil)
				},
				mockWordAligner: func(m *MockWordAligner) {
					m.On("Align", mock.Anything, &pb.AlignRequest{
						SourceText: "Text",
						TargetText: "Текст",
					}).Return(&pb.AlignResponse{
						Alignments: []*pb.AlignmentResult{},
					}, nil)
				},
				mockThrottler: func(m *MockThrottler) {
					m.On("Throttle").Times(2)
				},
				expectedResult: &domain.ChapterAlignNode{
					Id:              "img1",
					Title:           "Image Chapter",
					TranslatedTitle: "Глава с картинкой",
					Content: []domain.ContentNode{
						{
							Type: domain.ParagraphAlignNodeTypeParagraph,
							ParagraphAlignNode: &domain.ParagraphAlignNode{
								OriginalParagraph:   "Text",
								TranslatedParagraph: "Текст",
								AlignmentWords:      []domain.WordAlignNode{},
							},
						},
						{
							Type: domain.ParagraphAlignNodeTypeImage,
							ImageNode: &domain.ImageNode{
								ImageURL: "",
							},
						},
					},
				},
				expectedError: "",
			},
		*/
		{
			name: "Empty Chapter",
			chapter: epub_parser.FormattedChapter{
				Chapter: pamphlet.Chapter{
					ID:    "2",
					Title: "",
				},
				Content: []epub_parser.ContentItem{},
			},
			mockTranslator:  nil,
			mockWordAligner: nil,
			mockThrottler:   nil,
			expectedResult: &domain.ChapterAlignNode{
				Id:              "2",
				Title:           "",
				TranslatedTitle: "",
				Content:         []domain.ContentNode{},
			},
			expectedError: "",
		},
		{
			name: "Paragraph Translation Error",
			chapter: epub_parser.FormattedChapter{
				Chapter: pamphlet.Chapter{
					ID:    "3",
					Title: "Title",
				},
				Content: []epub_parser.ContentItem{
					{
						Type:     epub_parser.ContentTypeText,
						TextNode: &epub_parser.TextNode{Text: "Fail"},
					},
				},
			},
			mockTranslator: func(m *MockTranslator) {
				m.On("Translate", &translate.Input{
					Q:      "Fail",
					Source: domain.En,
					Target: domain.Ru,
				}).Return(nil, errors.New("translate error"))
			},
			mockWordAligner: nil,
			mockThrottler:   nil,
			expectedResult:  nil,
			expectedError:   "translate.logic.TranslateChapter: failed to translate paragraph",
		},
		{
			name: "Title Translation Error",
			chapter: epub_parser.FormattedChapter{
				Chapter: pamphlet.Chapter{
					ID:    "4",
					Title: "Fail Title",
				},
				Content: []epub_parser.ContentItem{
					{
						Type:     epub_parser.ContentTypeText,
						TextNode: &epub_parser.TextNode{Text: "Hello"},
					},
				},
			},
			mockTranslator: func(m *MockTranslator) {
				// Translate Paragraph (Success)
				m.On("Translate", &translate.Input{
					Q:      "Hello",
					Source: domain.En,
					Target: domain.Ru,
				}).Return(&translate.Output{TranslatedText: "Привет"}, nil)

				// Translate Title (Fail)
				m.On("Translate", &translate.Input{
					Q:      "Fail Title",
					Source: domain.En,
					Target: domain.Ru,
				}).Return(nil, errors.New("title error"))
			},
			mockWordAligner: func(m *MockWordAligner) {
				m.On("Align", mock.Anything, mock.Anything).Return(&pb.AlignResponse{}, nil)
			},
			mockThrottler: func(m *MockThrottler) {
				m.On("Throttle").Times(2)
			},
			expectedResult: nil,
			expectedError:  "translate.logic.TranslateChapter: failed to translate title: title error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewMockTranslator(t)
			wordAligner := NewMockWordAligner(t)
			throttler := NewMockThrottler(t)

			if tt.mockTranslator != nil {
				tt.mockTranslator(translator)
			}
			if tt.mockWordAligner != nil {
				tt.mockWordAligner(wordAligner)
			}
			if tt.mockThrottler != nil {
				tt.mockThrottler(throttler)
			}

			l := New(translator, wordAligner)
			res, err := l.TranslateChapter(context.Background(), tt.chapter, domain.En, domain.Ru, throttler, 1, nil)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, res)
			}
		})
	}
}
