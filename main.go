package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	// "regexp"
	"strings"
	"time"

	// "github.com/gocolly/colly/v2"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v2"
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
		article, err := parseFile(file)
		if err != nil {
			log.Printf("Error parsing file %s: %s\n", file, err)
			continue
		}
		articles = append(articles, article)
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

func parseFile(file string) (Article, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Article{}, err
	}

	// Split the file content into frontmatter and content
	frontmatter, content := splitFrontmatterAndContent(string(data))

	// Parse the frontmatter
	var article Article
	err = yaml.Unmarshal([]byte(frontmatter), &article)
	if err != nil {
		return Article{}, err
	}

	// Set default values for created and updated if not provided
	if article.Created.IsZero() {
		article.Created = time.Now()
	}
	if article.Updated.IsZero() {
		article.Updated = article.Created
	}

	// Convert markdown to HTML
	if strings.HasSuffix(file, ".md") {
		article.Content = string(blackfriday.Run([]byte(content)))
	} else {
		article.Content = content
	}

	// Extract resources from HTML
	article.Files = extractResources(article.Content)

	// Determine if the article is a page
	article.IsPage = contains(article.Tags, "PAGE")

	// Set the article path
	article.Path = filepath.Dir(file)

	return article, nil
}

func splitFrontmatterAndContent(data string) (string, string) {
	parts := strings.SplitN(data, "---", 3)
	if len(parts) != 3 {
		return "", data
	}
	return parts[1], parts[2]
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
    articleDir := filepath.Join(outputDir, strings.ReplaceAll(article.Title, " ", "-"))
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
