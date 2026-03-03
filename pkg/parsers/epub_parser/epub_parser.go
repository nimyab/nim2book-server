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
	TagsForRemove = map[string]struct{}{
		"code":   {},
		"script": {},
		"style":  {},
		"pre":    {},
		"a":      {},
		"strong": {},
		"br":     {},
		"hr":     {},
	}

	TagsForExtractTitle = []string{
		"h1",
		"h2",
		"h3",
		"title",
		"h4",
		"h5",
		"h6",
	}
)

type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
)

type ContentUnit struct {
	Type      ContentType
	ImageNode *ImageUnit
	TextNode  *TextUnit
}

type ImageUnit struct {
	File pamphlet.ZipFile
	Alt  string
}

type TextUnit struct {
	Text string
}

type FormattedChapter struct {
	PamphletChapterData pamphlet.Chapter
	Content             []ContentUnit
	CapterTitle         string
}

type ParsedData struct {
	FormattedChapter []FormattedChapter
	Cover            []byte
	Book             *pamphlet.Book
}

func Parse(data []byte) (*ParsedData, error) {
	const operation = "pkg.parsers.epub_parser.Parse"

	parser, err := pamphlet.OpenBytes(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer func() {
		if err := parser.Close(); err != nil {
			slog.Error("failed to close parser", slog.String("error", err.Error()), slog.String("operation", operation))
		}
	}()

	book := parser.GetBook()
	if len(book.Chapters) == 0 {
		return nil, fmt.Errorf("%s: no chapters found in the book", operation)
	}

	coverData, err := extractCover(book)
	if err != nil {
		slog.Error(err.Error())
	}

	formattedChapters := make([]FormattedChapter, 0, len(book.Chapters))
	for _, chapter := range book.Chapters {
		content, err := chapter.GetContent()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		items, err := extractContentFromHtml(book, content)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		if len(items) == 0 {
			continue
		}

		chapterTitle, err := extractChapterTitleFromHtml(content)
		if err != nil {
			slog.Warn(fmt.Sprintf("Chapter title not found for chapter with href %s: %v", chapter.Href, err))
			chapterTitle = chapter.Title
		}

		formattedChapters = append(formattedChapters, FormattedChapter{
			PamphletChapterData: chapter,
			Content:             items,
			CapterTitle:         chapterTitle,
		})
	}

	return &ParsedData{
		FormattedChapter: formattedChapters,
		Cover:            coverData,
		Book:             book,
	}, nil
}

func extractContentFromHtml(book *pamphlet.Book, xmlText string) ([]ContentUnit, error) {
	const operation = "pkg.parsers.epub_parser.extractContentFromHtml"

	doc, err := html.Parse(strings.NewReader(xmlText))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	var contentUnits []ContentUnit

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		// обработка параграфов
		if node.Type == html.ElementNode && node.Data == "p" {
			for c := node.FirstChild; c != nil; {
				next := c.NextSibling
				if _, ok := TagsForRemove[c.Data]; ok {
					node.RemoveChild(c)
				}
				c = next
			}

			paragraph := strings.TrimSpace(extractTextContent(node))
			paragraph = regexp.MustCompile(`\s+`).ReplaceAllString(paragraph, " ")
			if paragraph != "" {
				contentUnits = append(contentUnits, ContentUnit{
					Type:     ContentTypeText,
					TextNode: &TextUnit{Text: paragraph},
				})
			}
		}
		// обработка изображений
		if node.Type == html.ElementNode && (node.Data == "img" || node.Data == "image") {
			var href, alt string
			for _, attr := range node.Attr {
				switch attr.Key {
				case "src", "href", "xlink:href":
					href = attr.Val
				case "alt":
					alt = attr.Val
				}
			}
			if href != "" {
				imageFile, err := resolveImage(book, href)
				if err != nil {
					slog.Error("failed to resolve image", slog.String("href", href), slog.String("error", err.Error()))
				} else {
					contentUnits = append(contentUnits, ContentUnit{
						Type: ContentTypeImage,
						ImageNode: &ImageUnit{
							File: imageFile,
							Alt:  strings.TrimSpace(alt),
						},
					})
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return contentUnits, nil
}

// resolveImage пытаемся найти изображение в манифесте по src
func resolveImage(book *pamphlet.Book, src string) (pamphlet.ZipFile, error) {
	for _, item := range book.ManifestItems {
		if item.Href == src || item.RealPath == src || strings.HasSuffix(item.Href, src) || strings.HasSuffix(item.RealPath, src) {
			return item.ZipFile, nil
		}
	}

	return pamphlet.ZipFile{}, fmt.Errorf("image not found: %s", src)
}

// extractCover получаем обложку книги. Сначала пытаемся найти изображение на титульной странице, если не находим, то берем первое изображение из manifest
func extractCover(book *pamphlet.Book) ([]byte, error) {
	const operation = "pkg.parsers.epub_parser.extractCover"

	var coverItem *pamphlet.ManifestItem

	// Ищем обложку в 1 файле, это обычно титульная страница
	titlePageContent, err := book.Chapters[0].GetContent()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	doc, err := html.Parse(strings.NewReader(titlePageContent))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	for node := range doc.Descendants() {
		if node.Type == html.ElementNode && (node.Data == "img" || node.Data == "image") {
			for _, attr := range node.Attr {
				if attr.Key == "src" || attr.Key == "href" {
					for _, item := range book.ManifestItems {
						if item.Href == attr.Val || item.RealPath == attr.Val {
							coverItem = &item
							break
						}
					}
				}
				if coverItem != nil {
					break
				}
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

// extractChapterTitleFromHtml извлекает название главы из HTML-контента главы. Обычно название главы находится в теге <title>, но может быть и в других местах, поэтому функция ищет его рекурсивно по всему дереву HTML.
func extractChapterTitleFromHtml(xmlText string) (string, error) {
	const operation = "pkg.parsers.epub_parser.extractChapterTitleFromHtml"

	doc, err := html.Parse(strings.NewReader(xmlText))
	if err != nil {
		return "", fmt.Errorf("%s: %w", operation, err)
	}

	// Сначала ищем h1 потом h2 и так далее, по списку
	for _, titleTag := range TagsForExtractTitle {
		for node := range doc.Descendants() {
			if node.Type == html.ElementNode && node.Data == titleTag {
				title := strings.TrimSpace(extractTextContent(node))
				title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
				if title != "" {
					return title, nil
				}
			}
		}
	}

	return "", fmt.Errorf("chapter title not found")
}

// extractTextContent извлекает текст из узла HTML. Если узел является текстовым узлом, возвращается его данные. В противном случае рекурсивно вызывается для всех дочерних узлов и объединяются их результаты.
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
