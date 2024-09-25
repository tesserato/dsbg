package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	// defaultTitle      = "My Blog"
	// defaultContentDir = "content"
	// defaultOutputDir  = "public"
	defaultIndexName  = "index.html"
	// defaultDateFormat = "2006-01-02" // TODO look closer
)

type Article struct {
	Title        string
	Description  string
	Created      time.Time
	Updated      time.Time
	Tags         []string
	HtmlContent  string
	// Files        []string
	// IsPage       bool
	OriginalPath string
}

func (a Article) SaveHtml(outputDir string) error {
	// Create the page folder if it doesn't exist
	pageDir := filepath.Join(outputDir, a.OriginalPath)
	err := os.MkdirAll(pageDir, 0755)
	if err != nil {
		return err
	}

	// resources := extractResources(a.HtmlContent)


	// Sanitize the HTML content
	// p := bluemonday.UGCPolicy()
	// html = p.Sanitize(html)

	// Write the HTML content to the file
	filename := filepath.Join(pageDir, defaultIndexName)
	return os.WriteFile(filename, []byte(a.HtmlContent), 0644)
}

// func GenerateTagsHTML(tags []string) string {
// 	var html string
// 	for _, tag := range tags {
// 		html += fmt.Sprintf("<li>%s</li>\n", tag)
// 	}
// 	return html
// }

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
