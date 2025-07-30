package epub_parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/timsims/pamphlet"
	"golang.org/x/net/html"
)

var (
	TagsForRemove = map[string]bool{
		"code":   true,
		"a":      true,
		"strong": true,
		"pre":    true,
		"html":   true,
		"body":   true,
		"head":   true,
		"script": true,
		"style":  true,
		"h1":     true,
		"h2":     true,
		"h3":     true,
		"h4":     true,
		"h5":     true,
		"h6":     true,
		"img":    true,
		"br":     true,
		"hr":     true,
	}
)

type FormattedChapter struct {
	pamphlet.Chapter
	Paragraphs []string
}

func Parse(data []byte) (*pamphlet.Book, []FormattedChapter, error) {
	const operation = "pkg.parsers.epub_parser.Parse"

	parser, err := pamphlet.OpenBytes(data)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer parser.Close()
	book := parser.GetBook()
	formattedChapters := make([]FormattedChapter, 0, len(book.Chapters))
	for _, chapter := range book.Chapters {
		content, err := chapter.GetContent()
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", operation, err)
		}
		paragraphs, err := extractTextFromHtml(content)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", operation, err)
		}
		if len(paragraphs) == 0 {
			continue
		}
		formattedChapters = append(formattedChapters, FormattedChapter{
			Chapter:    chapter,
			Paragraphs: paragraphs,
		})
	}

	return book, formattedChapters, nil
}

func extractTextFromHtml(xmlText string) ([]string, error) {
	const operation = "pkg.parsers.epub_parser.extractTextFromHtml"

	doc, err := html.Parse(strings.NewReader(xmlText))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	var paragraphs []string
	for node := range doc.Descendants() {
		if node.Type == html.ElementNode && node.Data == "p" {
			for childrenP := range node.ChildNodes() {
				if TagsForRemove[childrenP.Data] {
					node.RemoveChild(childrenP)
				}
				paragraph := extractTextContent(node)
				paragraph = string(regexp.MustCompile(`\s+`).ReplaceAll([]byte(paragraph), []byte(" ")))
				if strings.TrimSpace(paragraph) != "" {
					paragraphs = append(paragraphs, strings.TrimSpace(paragraph))
				}
			}

		}
	}

	return paragraphs, nil
}

func extractTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var result strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result.WriteString(extractTextContent(c))
	}

	return result.String()
}
