package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	// "github.com/gocolly/colly/v2"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	// "github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
	// "gopkg.in/yaml.v2"
)

const (
	defaultTitle      = "My Blog"
	defaultContentDir = "content"
	defaultOutputDir  = "public"
	defaultIndexName  = "index.html"
	defaultDateFormat = "2006-01-02"
)

type Article struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Created     time.Time `yaml:"created"`
	Updated     time.Time `yaml:"updated"`
	Tags        []string  `yaml:"tags"`
	Content     string    `yaml:"-"`
	Files       []string  `yaml:"-"`
	IsPage      bool      `yaml:"-"`
	Path        string    `yaml:"-"`
}

func main() {
	// 1. Read all markdown and HTML files from the content folder
	files, err := getFiles(defaultContentDir, []string{".md", ".html"})
	if err != nil {
		log.Fatal(err)
	}

	// 2. Parse each file into an Article struct
	var articles []Article
	for _, file := range files {
		if strings.HasSuffix(file, ".md") {
			article, err := parseMarkdownFile(file)
			if err != nil {
				log.Printf("Error parsing file %s: %s\n", file, err)
				continue
			}
			articles = append(articles, article)

		} else if strings.HasSuffix(file, ".html") {
			article, err := parseHTMLFile(file)
			if err != nil {
				log.Printf("Error parsing file %s: %s\n", file, err)
				continue
			}
			articles = append(articles, article)
		}
	}

	// 3. Generate HTML for each article and page
	for _, article := range articles {
		if article.IsPage {
			err := generatePageHTML(article, defaultOutputDir)
			if err != nil {
				log.Printf("Error generating HTML for page %s: %s\n", article.Title, err)
			}
		} else {
			err := generateArticleHTML(article, defaultOutputDir)
			if err != nil {
				log.Printf("Error generating HTML for article %s: %s\n", article.Title, err)
			}
		}
	}

	// 4. Generate the index.html file
	err = generateIndexHTML(articles, defaultOutputDir, defaultIndexName)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Blog generated successfully!")
}

func getFiles(root string, extensions []string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			for _, ext := range extensions {
				if strings.HasSuffix(path, ext) {
					files = append(files, path)
					break
				}
			}
		}
		return nil
	})
	return files, err
}

func parseMarkdownFile(file string) (Article, error) {
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
	article.Content = content

	return article, nil
}

func parseHTMLFile(file string) (Article, error) {
	// Read the HTML file content
	data, err := os.ReadFile(file)
	if err != nil {
		return Article{}, fmt.Errorf("failed to read file %s: %w", file, err)
	}
	content := string(data)

	// Create an article and populate common fields
	article := Article{
		Content: content,
		Files:   extractResources(content),
		IsPage:  true, // Assume HTML files represent pages
		Path:    filepath.Dir(file),
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

func extractResources(htmlContent string) []string {
	var resources []string
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		log.Println("Error parsing HTML:", err)
		return resources // Return an empty slice on error
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Helper function to copy a file
func copyFile(src, dest string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dest, input, 0644)
}

func generateArticleHTML(article Article, outputDir string) error {
	// Create the article folder if it doesn't exist
	articleDir := filepath.Join(outputDir, strings.ReplaceAll(article.Path, " ", "_"))
	err := os.MkdirAll(articleDir, 0755)
	if err != nil {
		return err
	}

	// Copy resources to the article folder
	for _, file := range article.Files {
		srcPath := filepath.Join(article.Path, file)
		destPath := filepath.Join(articleDir, file)

		// Create destination directory if it doesn't exist
		err := os.MkdirAll(filepath.Dir(destPath), 0755)
		if err != nil {
			return err
		}

		err = copyFile(srcPath, destPath)
		if err != nil {
			return fmt.Errorf("failed to copy file %s: %w", file, err)
		}
	}

	// Generate the HTML content
	html := fmt.Sprintf(`
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
		article.Title,
		article.Title,
		article.Created.Format(defaultDateFormat),
		article.Updated.Format(defaultDateFormat),
		generateTagsHTML(article.Tags),
		article.Content,
	)

	// Sanitize the HTML content
	p := bluemonday.UGCPolicy()
	html = p.Sanitize(html)

	// Write the HTML content to the file
	filename := filepath.Join(articleDir, defaultIndexName)
	return os.WriteFile(filename, []byte(html), 0644)
}

func generatePageHTML(article Article, outputDir string) error {
	// Create the page folder if it doesn't exist
	pageDir := filepath.Join(outputDir, article.Path)
	err := os.MkdirAll(pageDir, 0755)
	if err != nil {
		return err
	}

	// Generate the HTML content
	html := fmt.Sprintf(`
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
		article.Title,
		article.Title,
		article.Content,
	)

	// Sanitize the HTML content
	p := bluemonday.UGCPolicy()
	html = p.Sanitize(html)

	// Write the HTML content to the file
	filename := filepath.Join(pageDir, defaultIndexName)
	return os.WriteFile(filename, []byte(html), 0644)
}

func generateTagsHTML(tags []string) string {
	var html string
	for _, tag := range tags {
		html += fmt.Sprintf("<li>%s</li>\n", tag)
	}
	return html
}

func generateIndexHTML(articles []Article, outputDir, filename string) error {
	// Generate the article list HTML
	var articleList string
	for _, article := range articles {
		if !article.IsPage {
			articleLink := strings.ReplaceAll(article.Title, " ", "-") + "/"
			articleList += fmt.Sprintf(`
			<li>
				<a href="%s">%s</a>
				<p>%s</p>
				<p>Created: %s</p>
				<p>Updated: %s</p>
				<ul>
					%s
				</ul>
			</li>
			`,
				articleLink,
				article.Title,
				article.Description,
				article.Created.Format(defaultDateFormat),
				article.Updated.Format(defaultDateFormat),
				generateTagsHTML(article.Tags),
			)
		}
	}

	// Generate the index.html content
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>%s</title>
		<link rel="stylesheet" href="/style.css">
	</head>
	<body>
		<h1>%s</h1>
		<ul>
			%s
		</ul>
	</body>
	</html>
	`,
		defaultTitle,
		defaultTitle,
		articleList,
	)

	// Write the HTML content to the file
	filePath := filepath.Join(outputDir, filename)
	return os.WriteFile(filePath, []byte(html), 0644)
}
