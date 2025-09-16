package epub_parser

import (
	"fmt"
	"log/slog"
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

func Parse(data []byte) (*pamphlet.Book, []FormattedChapter, []byte, error) {
	const operation = "pkg.parsers.epub_parser.Parse"

	parser, err := pamphlet.OpenBytes(data)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer parser.Close()

	book := parser.GetBook()

	coverData, err := extractCover(book)
	if err != nil {
		slog.Error(err.Error())
	}

	formattedChapters := make([]FormattedChapter, 0, len(book.Chapters))
	for _, chapter := range book.Chapters {
		content, err := chapter.GetContent()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s: %w", operation, err)
		}
		paragraphs, err := extractTextFromHtml(content)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s: %w", operation, err)
		}
		if len(paragraphs) == 0 {
			continue
		}
		formattedChapters = append(formattedChapters, FormattedChapter{
			Chapter:    chapter,
			Paragraphs: paragraphs,
		})
	}

	return book, formattedChapters, coverData, nil
}

func extractTextFromHtml(xmlText string) ([]string, error) {
	const operation = "pkg.parsers.epub_parser.extractTextFromHtml"

	doc, err := html.Parse(strings.NewReader(xmlText))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	var paragraphs []string
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "p" {
			for c := node.FirstChild; c != nil; {
				next := c.NextSibling
				if TagsForRemove[c.Data] {
					node.RemoveChild(c)
				}
				c = next
			}
			paragraph := extractTextContent(node)
			paragraph = string(regexp.MustCompile(`\s+`).ReplaceAll([]byte(paragraph), []byte(" ")))
			if strings.TrimSpace(paragraph) != "" {
				paragraphs = append(paragraphs, strings.TrimSpace(paragraph))
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
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

func extractCover(book *pamphlet.Book) ([]byte, error) {
	const operation = "pkg.parsers.epub_parser.extractCover"

	var coverItem *pamphlet.ManifestItem

	for _, item := range book.ManifestItems {
		if strings.HasPrefix(item.MediaType, "image/") {
			// Ищем файлы с "cover" в имени или пути
			if strings.Contains(strings.ToLower(item.Href), "cover") ||
				strings.Contains(strings.ToLower(item.RealPath), "cover") {
				coverItem = &item
				break
			}
		}
	}

	// Если не нашли обложку по имени, берем первое изображение
	if coverItem == nil {
		for _, item := range book.ManifestItems {
			if strings.HasPrefix(item.MediaType, "image/") {
				coverItem = &item
				break
			}
		}
	}

	if coverItem == nil {
		return nil, fmt.Errorf("cover not found")
	}

	data, err := coverItem.GetRawContent()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return data, nil
}
