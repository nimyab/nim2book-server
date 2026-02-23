package epub_parser

import (
	"fmt"
	"log/slog"
	"path/filepath"
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
		"br":     true,
		"hr":     true,
	}
)

type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
)

type ContentItem struct {
	Type      ContentType
	ImageNode *ImageNode
	TextNode  *TextNode
}

type ImageNode struct {
	ImageFile pamphlet.ZipFile
}

type TextNode struct {
	Text string
}

type FormattedChapter struct {
	pamphlet.Chapter
	Content []ContentItem
}

func Parse(data []byte) (*pamphlet.Book, []FormattedChapter, []byte, error) {
	const operation = "pkg.parsers.epub_parser.Parse"

	parser, err := pamphlet.OpenBytes(data)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", operation, err)
	}
	// We do NOT close parser here, because ImageNode needs access to zip files later.
	// defer parser.Close()

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
		items, err := extractContentFromHtml(book, content)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s: %w", operation, err)
		}
		if len(items) == 0 {
			continue
		}
		formattedChapters = append(formattedChapters, FormattedChapter{
			Chapter: chapter,
			Content: items,
		})
	}

	return book, formattedChapters, coverData, nil
}

func isBlockElement(n string) bool {
	switch n {
	case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6", "li", "blockquote", "section", "article", "header", "footer":
		return true
	}
	return false
}

func extractContentFromHtml(book *pamphlet.Book, xmlText string) ([]ContentItem, error) {
	const operation = "pkg.parsers.epub_parser.extractContentFromHtml"

	doc, err := html.Parse(strings.NewReader(xmlText))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	var items []ContentItem
	var currentTextBuilder strings.Builder

	flushText := func() {
		if currentTextBuilder.Len() > 0 {
			text := currentTextBuilder.String()
			// Clean text
			text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
			if strings.TrimSpace(text) != "" {
				items = append(items, ContentItem{
					Type: ContentTypeText,
					TextNode: &TextNode{
						Text: strings.TrimSpace(text),
					},
				})
			}
			currentTextBuilder.Reset()
		}
	}

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "img" {
			flushText()
			processImage(book, node, &items)
			return
		}

		if node.Type == html.TextNode {
			currentTextBuilder.WriteString(node.Data)
			return
		}

		// Recurse for other nodes
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && TagsForRemove[c.Data] {
				// Special case: allow html and body to be traversed
				if c.Data != "html" && c.Data != "body" {
					continue
				}
			}
			traverse(c)
		}

		if node.Type == html.ElementNode && isBlockElement(node.Data) {
			flushText()
		}
	}
	traverse(doc)

	flushText()

	return items, nil
}

func processImage(book *pamphlet.Book, imgNode *html.Node, items *[]ContentItem) {
	// Extract src
	var src string
	for _, attr := range imgNode.Attr {
		if attr.Key == "src" || attr.Key == "href" {
			src = attr.Val
			break
		}
	}
	if src == "" {
		return
	}

	// Resolve image file
	file, err := resolveImage(book, src)
	if err != nil {
		slog.Error("failed to resolve image", slog.String("src", src), slog.String("error", err.Error()))
		return
	}

	*items = append(*items, ContentItem{
		Type:      ContentTypeImage,
		ImageNode: &ImageNode{ImageFile: file},
	})
}

func resolveImage(book *pamphlet.Book, src string) (pamphlet.ZipFile, error) {
	// Try exact match first (if src is full path)
	// Note: href in manifest might be URI encoded, src might be too.
	// We assume pamphlet handles some of this or we do basic matching.

	// Normalize src?
	// If src is "../Images/foo.jpg", and manifest href is "OEBPS/Images/foo.jpg".
	// Simplest strategy: Match by base filename.
	base := filepath.Base(src)

	for _, item := range book.ManifestItems {
		// Try exact match
		if item.Href == src || item.RealPath == src {
			return item.ZipFile, nil
		}
		// Try base match
		if filepath.Base(item.Href) == base {
			return item.ZipFile, nil
		}
	}

	return pamphlet.ZipFile{}, fmt.Errorf("image not found: %s", src)
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
