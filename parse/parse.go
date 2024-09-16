package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
	"golang.org/x/net/html"
)
// type ArticleMarkdownMetadata struct {
// 	Title         string   `yaml:"title"`
// 	Description   string   `yaml:"description"`
// 	CreatedString string   `yaml:"created"`
// 	UpdatedString string   `yaml:"updated"`
// 	Tags          []string `yaml:"tags"`
// }

func MarkdownFile(file string) (Article, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Article{}, err
	}

	markdown := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAttribute(),
			parser.WithAutoHeadingID(),
		),
		goldmark.WithExtensions(
			&frontmatter.Extender{},
		),
	)

	// Create a context to store frontmatter
	context := parser.NewContext()

	// Parse the Markdown file, storing frontmatter in the context
	var buf strings.Builder
	if err := markdown.Convert(data, &buf, parser.WithContext(context)); err != nil {
		panic(err)
	}
	content := buf.String()

	// Retrieve frontmatter from the context
	var rawArticle map[string]interface{}
	d := frontmatter.Get(context)
	if d != nil {
		if err := d.Decode(&rawArticle); err != nil {
			return Article{}, fmt.Errorf("failed to unmarshal frontmatter: %w", err)
		}
	}

	var article Article

	// 1. Get info from rawArticle and add to article
	if title, ok := rawArticle["title"].(string); ok {
		article.Title = title
	}
	if description, ok := rawArticle["description"].(string); ok {
		article.Description = description
	}
	if created, ok := rawArticle["created"].(string); ok {
		for _, layout := range []string{"2006-01-02", "02/01/2006", "01/02/2006"} {
			t, err := time.Parse(layout, created)
			if err == nil {
				article.Created = t
				break
			}
		}
	}
	if updated, ok := rawArticle["updated"].(string); ok {
		t, err := time.Parse(defaultDateFormat, updated)
		if err != nil {
			return Article{}, fmt.Errorf("invalid 'updated' date format: %w", err)
		}
		article.Updated = t
	}
	if tags, ok := rawArticle["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if strTag, ok := tag.(string); ok {
				article.Tags = append(article.Tags, strTag)
			}
		}
	}

	// 2. Set Created and Updated to file dates if not provided in frontmatter
	fileInfo, err := os.Stat(file)
	if err != nil {
		return Article{}, fmt.Errorf("failed to get file info: %w", err)
	}
	if article.Created.IsZero() {
		article.Created = fileInfo.ModTime() // Use file modification time
	}
	if article.Updated.IsZero() {
		article.Updated = fileInfo.ModTime() // Use file modification time
	}

	// 3. Default title to filename if not provided
	if article.Title == "" {
		article.Title = strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	}
	// Extract resources from HTML
	article.Files = extractResources(content) // Pass content here, not article.Content

	// Determine if the article is a page
	article.IsPage = contains(article.Tags, "PAGE")

	// Set the article path
	article.Path = filepath.Dir(file)
	// Set the HTML content (Goldmark already converted it)
	article.HtmlContent = content

	return article, nil
}

func HTMLFile(file string) (Article, error) {
	// Read the HTML file content
	data, err := os.ReadFile(file)
	if err != nil {
		return Article{}, fmt.Errorf("failed to read file %s: %w", file, err)
	}
	content := string(data)

	// Create an article and populate common fields
	article := Article{
		HtmlContent: content,
		Files:       extractResources(content),
		IsPage:      true, // Assume HTML files represent pages
		Path:        filepath.Dir(file),
	}

	// Extract title from the first H1 tag
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return Article{}, fmt.Errorf("failed to parse HTML: %w", err)
	}
	if h1 := findFirstElement(doc, "h1"); h1 != nil {
		article.Title = getTextContent(h1)
	}

	// If no H1 tag, default title to filename
	if article.Title == "" {
		article.Title = strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	}

	// Use file modification time for Created and Updated
	fileInfo, err := os.Stat(file)
	if err != nil {
		return Article{}, fmt.Errorf("failed to get file info: %w", err)
	}
	article.Created = fileInfo.ModTime()
	article.Updated = fileInfo.ModTime()

	return article, nil
}

func extractResources(htmlContent string) []string {
	var resources []string
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		fmt.Errorf("error parsing HTML: %w", err)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "img" || n.Data == "script" || n.Data == "link" {
				for _, attr := range n.Attr {
					if attr.Key == "src" || attr.Key == "href" {
						resources = append(resources, attr.Val)
						break // Assuming only one relevant attribute per tag
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return resources
}

// Helper function to find the first occurrence of an element by tag name
func findFirstElement(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirstElement(c, tag); found != nil {
			return found
		}
	}
	return nil
}

// Helper function to extract text content from an HTML node
func getTextContent(n *html.Node) string {
	var text string
	if n.Type == html.TextNode {
		text += n.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getTextContent(c)
	}
	return strings.TrimSpace(text)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}