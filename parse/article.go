package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// defaultTitle      = "My Blog"
	// defaultContentDir = "content"
	// defaultOutputDir  = "public"
	defaultIndexName  = "index.html"
	defaultDateFormat = "2006-01-02" // TODO look closer
)

type Article struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Created     time.Time `yaml:"created"`
	Updated     time.Time `yaml:"updated"`
	Tags        []string  `yaml:"tags"`
	HtmlContent string    `yaml:"-"`
	Files       []string  `yaml:"-"`
	IsPage      bool      `yaml:"-"`
	Path        string    `yaml:"-"`
}

func (a Article) SaveHtml(outputDir string) error {
	// Create the page folder if it doesn't exist
	pageDir := filepath.Join(outputDir, a.Path)
	err := os.MkdirAll(pageDir, 0755)
	if err != nil {
		return err
	}

	// Generate the HTML content
	html := ""
	if a.IsPage {
		html = fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>%s</title>
			<link rel="stylesheet" href="/style.css">
		</head>
		<body>
			<article>
				<h1>%s</h1>
				%s
			</article>
		</body>
		</html>
		`,
			a.Title,
			a.Title,
			a.HtmlContent,
		)
	} else {
		html = fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>%s</title>
			<link rel="stylesheet" href="/style.css">
		</head>
		<body>
			<article>
				<h1>%s</h1>
				<p>Created: %s</p>
				<p>Updated: %s</p>
				<ul>
					%s
				</ul>
				%s
			</article>
		</body>
		</html>
		`,
			a.Title,
			a.Title,
			a.Created.Format(defaultDateFormat),
			a.Updated.Format(defaultDateFormat),
			GenerateTagsHTML(a.Tags),
			a.HtmlContent,
		)

	}

	// Sanitize the HTML content
	// p := bluemonday.UGCPolicy()
	// html = p.Sanitize(html)

	// Write the HTML content to the file
	filename := filepath.Join(pageDir, defaultIndexName)
	return os.WriteFile(filename, []byte(html), 0644)
}

func GenerateTagsHTML(tags []string) string {
	var html string
	for _, tag := range tags {
		html += fmt.Sprintf("<li>%s</li>\n", tag)
	}
	return html
}
