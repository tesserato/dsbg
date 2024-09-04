package main

import (
	"fmt"
	// "io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	// "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v2"
)

type Article struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Created     time.Time `yaml:"created"`
	Updated     time.Time `yaml:"updated"`
	Tags        []string  `yaml:"tags"`
	Content     string    `yaml:"-"`
}

func main() {
	// 1. Read all markdown files from the content folder
	files, err := filepath.Glob("content/*.md")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Parse each markdown file into an Article struct
	var articles []Article
	for _, file := range files {
		article, err := parseArticle(file)
		if err != nil {
			log.Println("Error parsing article:", err)
			continue
		}
		articles = append(articles, article)
	}

	// 3. Generate HTML for each article
	for _, article := range articles {
		err := generateArticleHTML(article)
		if err != nil {
			log.Println("Error generating article HTML:", err)
			continue
		}
	}

	// 4. Generate the index.html file
	err = generateIndexHTML(articles)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Blog generated successfully!")
}

func parseArticle(file string) (Article, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Article{}, err
	}

	// Split the file content into frontmatter and markdown
	frontmatter, content := splitFrontmatterAndContent(string(data))

	// Parse the frontmatter
	var article Article
	err = yaml.Unmarshal([]byte(frontmatter), &article)
	if err != nil {
		return Article{}, err
	}

	// Convert markdown to HTML
	article.Content = string(blackfriday.Run([]byte(content)))

	// Set default values for created and updated if not provided
	if article.Created.IsZero() {
		article.Created = time.Now()
	}
	if article.Updated.IsZero() {
		article.Updated = article.Created
	}

	return article, nil
}

func splitFrontmatterAndContent(data string) (string, string) {
	parts := strings.SplitN(data, "---", 3)
	if len(parts) != 3 {
		return "", data
	}
	return parts[1], parts[2]
}

func generateArticleHTML(article Article) error {
	// Create the article folder if it doesn't exist
	articleDir := fmt.Sprintf("public/%s", strings.ReplaceAll(article.Title, " ", "-"))
	err := os.MkdirAll(articleDir, 0755)
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
	</head>
	<body>
		<h1>%s</h1>
		<p>Created: %s</p>
		<p>Updated: %s</p>
		<ul>
			%s
		</ul>
		%s
	</body>
	</html>
	`,
		article.Title,
		article.Title,
		article.Created.Format("2006-01-02"),
		article.Updated.Format("2006-01-02"),
		generateTagsHTML(article.Tags),
		article.Content,
	)

	// Sanitize the HTML content
	p := bluemonday.UGCPolicy()
	html = p.Sanitize(html)

	// Write the HTML content to the file
	filename := fmt.Sprintf("%s/index.html", articleDir)
	return os.WriteFile(filename, []byte(html), 0644)
}

func generateTagsHTML(tags []string) string {
	var html string
	for _, tag := range tags {
		html += fmt.Sprintf("<li>%s</li>\n", tag)
	}
	return html
}

func generateIndexHTML(articles []Article) error {
	// Generate the article list HTML
	var articleList string
	for _, article := range articles {
		articleList += fmt.Sprintf(`
		<li>
			<a href="%s">%s</a>
			<p>%s</p>
		</li>
		`,
			fmt.Sprintf("/%s/", strings.ReplaceAll(article.Title, " ", "-")),
			article.Title,
			article.Description,
		)
	}

	// Generate the index.html content
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>My Blog</title>
	</head>
	<body>
		<h1>My Blog</h1>
		<ul>
			%s
		</ul>
	</body>
	</html>
	`,
		articleList,
	)

	// Write the HTML content to the file
	return os.WriteFile("public/index.html", []byte(html), 0644)
}