package main

import (
	"dsbg/parse"
	"log"
	"os"
	"strings"
)

func main() {
	settings := parse.NewSettings()

	// 1. Read all markdown and HTML files from the content folder
	files, err := parse.GetPaths(settings.InputDirectory, []string{".md", ".html"})
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

			links := parse.CopyHtmlResources(settings, path, article.HtmlContent)

			article = parse.FormatMarkdown(article, links, settings)

			articles = append(articles, article)

			os.WriteFile(links.ToSave, []byte(article.HtmlContent), 0644)

		} else if strings.HasSuffix(lowerCasePath, ".html") {
			article, err := parse.HTMLFile(path)
			if err != nil {
				log.Printf("Error parsing file %s: %s\n", path, err)
				continue
			}
			links := parse.CopyHtmlResources(settings, path, article.HtmlContent)

			articles = append(articles, article)

			os.WriteFile(links.ToSave, []byte(article.HtmlContent), 0644)

		}
	}

	// for i, article := range articles {
	// }

	// 4. Generate the index.html file
	err = parse.GenerateHtmlIndex(articles, settings)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Blog generated successfully!")
}
