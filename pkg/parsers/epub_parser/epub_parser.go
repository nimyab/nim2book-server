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
		"a":      {},
		"strong": {},
		"pre":    {},
		"html":   {},
		"body":   {},
		"head":   {},
		"script": {},
		"style":  {},
		"h1":     {},
		"h2":     {},
		"h3":     {},
		"h4":     {},
		"h5":     {},
		"h6":     {},
		"img":    {},
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

type ImageData struct {
	Href string // ссылка на изображение в ManifestItem
	Alt  string // alt текст из HTML
}

type ContentUnit struct {
	Content   *string    // для параграфа
	ImageData *ImageData // для изображения
}

type FormattedChapter struct {
	pamphlet.Chapter
	ContentUnits []ContentUnit
	ChapterTitle string
}

type Book struct {
	PamphletBook *pamphlet.Book
	Chapters     []FormattedChapter
	Cover        []byte
}

func Parse(data []byte) (*Book, error) {
	const operation = "pkg.parsers.epub_parser.Parse"

	parser, err := pamphlet.OpenBytes(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer parser.Close()

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
		chapterTitle, err := extractChapterTitleFromHtml(content)
		if err != nil {
			slog.Warn(fmt.Sprintf("Chapter title not found for chapter with href %s: %v", chapter.Href, err))
			chapterTitle = chapter.Title
		}
		contentUnits, err := extractContentFromHtml(content)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}
		if len(contentUnits) == 0 {
			continue
		}
		formattedChapters = append(formattedChapters, FormattedChapter{
			Chapter:      chapter,
			ContentUnits: contentUnits,
			ChapterTitle: chapterTitle,
		})
	}

	resultBook := &Book{
		PamphletBook: book,
		Chapters:     formattedChapters,
		Cover:        coverData,
	}

	return resultBook, nil
}

// GetContentUnitsWithS3Paths возвращает главу с путями к изображениям в S3
func (b *Book) GetContentUnitsWithS3Paths(chapter *FormattedChapter, imagePathsMap map[string]string) []ContentUnit {
	result := make([]ContentUnit, len(chapter.ContentUnits))
	for i, unit := range chapter.ContentUnits {
		if unit.Content != nil {
			// Это параграф
			result[i] = unit
		} else if unit.ImageData != nil {
			// Это изображение - заменяем href на S3 путь если есть
			href := unit.ImageData.Href
			if s3Path, ok := imagePathsMap[href]; ok {
				result[i] = ContentUnit{
					ImageData: &ImageData{
						Href: s3Path,
						Alt:  unit.ImageData.Alt,
					},
				}
			} else {
				result[i] = unit
			}
		}
	}
	return result
}

func extractContentFromHtml(xmlText string) ([]ContentUnit, error) {
	const operation = "pkg.parsers.epub_parser.extractContentFromHtml"

	doc, err := html.Parse(strings.NewReader(xmlText))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	var contentUnits []ContentUnit
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
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
					Content: &paragraph,
				})
			}
		}
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
				contentUnits = append(contentUnits, ContentUnit{
					ImageData: &ImageData{
						Href: href,
						Alt:  strings.TrimSpace(alt),
					},
				})
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return contentUnits, nil
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

// findManifestItem ищет элемент в манифесте по href или realPath
func (b *Book) findManifestItem(href string) *pamphlet.ManifestItem {
	for i := range b.PamphletBook.ManifestItems {
		item := &b.PamphletBook.ManifestItems[i]
		if item.Href == href || item.RealPath == href || strings.HasSuffix(item.Href, href) || strings.HasSuffix(item.RealPath, href) {
			return item
		}
	}
	return nil
}

// GetImageData возвращает бинарные данные изображения по href. Используется для ленивой загрузки изображений.
func (b *Book) GetImageData(href string) ([]byte, string, error) {
	const operation = "pkg.parsers.epub_parser.Book.GetImageData"

	if href == "" {
		return nil, "", fmt.Errorf("%s: empty href", operation)
	}

	item := b.findManifestItem(href)
	if item == nil {
		return nil, "", fmt.Errorf("%s: manifest item not found for href: %s", operation, href)
	}

	data, err := item.GetRawContent()
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", operation, err)
	}

	return data, item.MediaType, nil
}

// GetParagraphs возвращает только текстовые параграфы из главы
func (fc *FormattedChapter) GetParagraphs() []string {
	paragraphs := make([]string, 0)
	for _, unit := range fc.ContentUnits {
		if unit.Content != nil {
			paragraphs = append(paragraphs, *unit.Content)
		}
	}
	return paragraphs
}

// GetImages возвращает только изображения из главы
func (fc *FormattedChapter) GetImages() []*ImageData {
	images := make([]*ImageData, 0)
	for _, unit := range fc.ContentUnits {
		if unit.ImageData != nil {
			images = append(images, unit.ImageData)
		}
	}
	return images
}
