package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"dsbg/parse"
)

func main() {
	settings := parse.NewSettings()

	// 1. Read all markdown and HTML files from the content folder
	files, err := getFiles(settings.InputDirectory, []string{".md", ".html"})
	if err != nil {
		log.Fatal(err)
	}

	// 2. Parse each file into an Article struct
	var articles []parse.Article
	for _, path := range files {
		lowerCasePath := strings.ToLower(path)
		if strings.HasSuffix(lowerCasePath, ".md") {
			article, err := parse.MarkdownFile(path)
			if err != nil {
				log.Printf("Error parsing file %s: %s\n", path, err)
				continue
			}
			articles = append(articles, article)

		} else if strings.HasSuffix(lowerCasePath, ".html") {
			article, err := parse.HTMLFile(path)
			if err != nil {
				log.Printf("Error parsing file %s: %s\n", path, err)
				continue
			}
			articles = append(articles, article)
		}
	}

	// 3. Generate HTML for each article and page
	for _, article := range articles {
		article.SaveHtml(settings.OutputDirectory)
	}

	// 4. Generate the index.html file
	err = generateIndexHTML(articles, settings)
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

// Helper function to copy a file
func copyFile(src, dest string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dest, input, 0644)
}

func generateIndexHTML(articles []parse.Article, settings parse.Settings) error {
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
				article.Created.Format(settings.DateFormat),
				article.Updated.Format(settings.DateFormat),
				parse.GenerateTagsHTML(article.Tags),
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
		settings.Title,
		settings.Title,
		articleList,
	)

	// Write the HTML content to the file
	filePath := filepath.Join(settings.OutputDirectory, settings.IndexName)
	return os.WriteFile(filePath, []byte(html), 0644)
}
