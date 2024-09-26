package main

import (
	"bytes"
	"dsbg/parse"
	"fmt"
	"html"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
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

			links := copyResources(settings, path, article.HtmlContent)

			article = parse.FormatMarkdown(article, links)

			articles = append(articles, article)

			os.WriteFile(links.ToSave, []byte(article.HtmlContent), 0644)

		} else if strings.HasSuffix(lowerCasePath, ".html") {
			article, err := parse.HTMLFile(path)
			if err != nil {
				log.Printf("Error parsing file %s: %s\n", path, err)
				continue
			}
			links := copyResources(settings, path, article.HtmlContent)

			articles = append(articles, article)

			os.WriteFile(links.ToSave, []byte(article.HtmlContent), 0644)

		}
	}

	// for i, article := range articles {
	// }

	// 4. Generate the index.html file
	err = generateIndexHTML(articles, settings)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Blog generated successfully!")
}

func copyResources(settings parse.Settings, originalArticlePath string, htmlContent string) parse.Links {
	relPath, err := filepath.Rel(settings.InputDirectory, originalArticlePath)
	if err != nil {
		panic(err)
	}
	pageDir := filepath.Join(settings.OutputDirectory, relPath)
	pageDir = strings.TrimSuffix(pageDir, filepath.Ext(pageDir))

	err = os.MkdirAll(pageDir, 0755)
	if err != nil {
		panic(err)
	}
	// name := filepath.Base(article.OriginalPath)
	savePath := filepath.Join(pageDir, settings.IndexName)

	// os.WriteFile(savePath, []byte(article.HtmlContent), 0644)

	link, err := filepath.Rel(settings.OutputDirectory, savePath)
	if err != nil {
		panic(err)
	}
	link = html.EscapeString(link)

	// articles[i].LinkToSelf = link

	fmt.Printf("inputDir: %s\norigPath: %s\nrelPath: %s\npageDir: %s\nfilename: %s\nlink: %s\n\n", settings.InputDirectory, originalArticlePath, relPath, pageDir, savePath, link)
	articleOrigFolderPath := filepath.Dir(originalArticlePath)
	articleDestFolderPath := filepath.Dir(savePath)
	for _, resourceOrigRelPath := range parse.ExtractResources(htmlContent) {
		resourceOrigPath := filepath.Join(articleOrigFolderPath, resourceOrigRelPath)
		resourceDestPath := filepath.Join(articleDestFolderPath, resourceOrigRelPath)
		fmt.Printf("  resourceOrigPath: %s\n  resourceDestPath: %s\n\n", resourceOrigPath, resourceDestPath)

		// resourceDestPath := filepath.Join(pageDir, resourceRelPath)
		// copyFile(resourceOrigPath, resourceDestPath)
	}

	staticBaseLink, err := filepath.Rel(pageDir, settings.OutputDirectory)
	if err != nil {
		panic(err)
	}

	return parse.Links{ToSelf: link, ToCss: staticBaseLink + "/style.css", ToJs: staticBaseLink + "/script.js", ToSave: savePath}
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

var htmlIndexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Settings.Title}}</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <h1>{{.Settings.Title}}</h1>
    {{range .ArticleList}}
        <div class="detail">
            <div class="headline">
                <a href="{{.LinkToSelf}}">
                    <h2>{{.Title}}</h2>
                </a>
                <div class="info">
                    <div class="tags">
                        {{range .Tags}}
                            <button class="on">{{.}}</button>
                        {{end}}
                    </div>
                    <h4 class="date">⋆ {{.Created.Format "2006/01/02"}}</h4>
                    <h4 class="date">♰ {{.Updated.Format "2006/01/02"}}</h4>
                </div>
            </div>
            <p class="description">{{.Description}}</p>
        </div>
    {{end}}
    <script src="index.js" async defer></script>
    <div class="giscus"></div>
</body>
</html>
`

func generateIndexHTML(articles []parse.Article, settings parse.Settings) error {
	// Generate the article list HTML
	var allTags []string
	var pageList []parse.Article
	var articleList []parse.Article
	for _, article := range articles {
		if slices.Contains(article.Tags, "PAGE") {
			pageList = append(pageList, article)
			// allTags = append(allTags, article.Tags...) TODO implement
		} else {
			allTags = append(allTags, article.Tags...)
			articleList = append(articleList, article)
		}
	}

	funcMap := template.FuncMap{
		"makeLink": func(title string) string {
			return strings.ReplaceAll(strings.ToLower(title), " ", "-") + "/"
		},
		"stringsJoin":    strings.Join,
		"slicesContains": slices.Contains[[]string],
	}
	tmpl, err := template.New("index.html").Funcs(funcMap).Parse(htmlIndexTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	var tp bytes.Buffer
	err = tmpl.Execute(&tp, struct {
		AllTags     []string
		PageList    []parse.Article
		ArticleList []parse.Article
		Settings    parse.Settings
	}{allTags, pageList, articleList, settings})
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	// Write the HTML content to the file
	filePath := filepath.Join(settings.OutputDirectory, settings.IndexName)
	return os.WriteFile(filePath, tp.Bytes(), 0644)
}
