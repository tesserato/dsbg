package main

import (
	"dsbg/parse"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"strings"
)

//go:embed assets/*
var assets embed.FS

func main() {
	// Parse command line arguments
	title := flag.String("title", "Blog", "The Title of the blog")
	inputDirectory := flag.String("input-dir", "content", "Path to the directory that holds the source files")
	outputDirectory := flag.String("output-dir", "public", "Path to the directory where the output files will be saved")
	dateFormat := flag.String("date-format", "2006-01-02", "Date format")
	indexName := flag.String("index-name", "index.html", "Name of the index files")
	pathToAdditionalElementsTop := flag.String("path-to-additional-elements-top", "", "Path to a file with additional elements (basically scripts) to be placed at the top of the HTML outputs")
	pathToAdditionalElemensBottom := flag.String("path-to-additional-elements-bottom", "", "Path to a file with additional elements (basically scripts) to be placed at the bottom of the HTML outputs")
	showHelp := flag.Bool("help", false, "Show help message")

	flag.Parse()

	// Check if help flag is set, if so, print help message and exit
	if *showHelp {
		fmt.Println("Usage: dsbg [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		return
	}

	// Parse command line arguments into settings struct
	settings := parse.Settings{
		Title:           *title,
		InputDirectory:  *inputDirectory,
		OutputDirectory: *outputDirectory,
		DateFormat:      *dateFormat,
		IndexName:       *indexName,
	}

	if *pathToAdditionalElementsTop != "" {
		additionalElementsTop, err := os.ReadFile(*pathToAdditionalElementsTop)
		if err != nil {
			log.Fatal(err)
		}

		settings.AdditionalElementsTop = template.HTML(additionalElementsTop)
	}

	if *pathToAdditionalElemensBottom != "" {
		additionalElemensBottom, err := os.ReadFile(*pathToAdditionalElemensBottom)
		if err != nil {
			log.Fatal(err)
		}
		settings.AdditionalElemensBottom = template.HTML(additionalElemensBottom)
	}

	// Read all markdown and HTML files from the content folder
	files, err := parse.GetPaths(settings.InputDirectory, []string{".md", ".html"})
	if err != nil {
		log.Fatal(err)
	}

	// Parse each file into an Article struct and save it
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

	// Generate the index.html file
	err = parse.GenerateHtmlIndex(articles, settings)
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve and save the css file
	css, err := assets.ReadFile("assets/style.css")
	if err != nil {
		log.Fatal(err)
	}
	pathToSaveCss := path.Join(settings.OutputDirectory, "style.css")
	os.WriteFile(pathToSaveCss, css, 0644)

	// Retrieve and save the js file
	js, err := assets.ReadFile("assets/script.js")
	if err != nil {
		log.Fatal(err)
	}
	pathToSaveJs := path.Join(settings.OutputDirectory, "script.js")
	os.WriteFile(pathToSaveJs, js, 0644)

	log.Println("Blog generated successfully!")
}
