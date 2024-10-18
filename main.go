package main

import (
	"dsbg/parse"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	// "github.com/k3a/html2text"
)

//go:embed assets/*
var assets embed.FS

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	var settings parse.Settings

	flag.StringVar(&settings.Title, "title", "Blog", "The Title of the blog")
	flag.StringVar(&settings.Description, "description", "This is my blog", "The description of the blog")
	flag.StringVar(&settings.InputDirectory, "input-path", "content", "Path to the directory that holds the source files")
	flag.StringVar(&settings.OutputDirectory, "output-path", "public", "Path to the directory where the output files will be saved")
	flag.StringVar(&settings.DateFormat, "date-format", "2006 01 02", "Date format")
	flag.StringVar(&settings.IndexName, "index-name", "index.html", "Name of the index files")
	flag.StringVar(&settings.PathToCustomCss, "css-path", "", "Path to a file with custom css")
	flag.StringVar(&settings.PathToCustomJs, "js-path", "", "Path to a file with custom js")
	flag.StringVar(&settings.PathToCustomFavicon, "favicon-path", "", "Path to a file with custom favicon")
	flag.BoolVar(&settings.DoNotExtractTagsFromPaths, "ignore-tags-from-paths", false, "Do not extract tags from path")
	flag.BoolVar(&settings.DoNotRemoveDateFromPaths, "keep-date-on-paths", false, "Do not remove date from path")
	styleString := flag.String("style", "default", "Style to be used")
	pathToAdditionalElementsTop := flag.String("elements-top", "", "Path to a file with additional HTML elements (basically scripts) to be placed at the top of the HTML outputs")
	pathToAdditionalElemensBottom := flag.String("elements-bottom", "", "Path to a file with additional HTML elements (basically scripts) to be placed at the bottom of the HTML outputs")
	showHelp := flag.Bool("help", false, "Show help message")
	watch := flag.Bool("watch", false, "Watch for changes and rebuild")
	// generate md template
	createTemplate := flag.Bool("template", false, "Create a markdown template with frontmatter fields. title, output-dir (defaults to current dir in this case) and date-format")

	flag.Parse()

	if *showHelp {
		fmt.Println("Usage: dsbg [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		return
	}

	if *createTemplate {
		tmpl, err := template.New("frontmatter").Parse(parse.FrontMatterTemplate)
		if err != nil {
			log.Fatalf("Error parsing template: %v", err)
		}

		formattedDate := time.Now().Format(settings.DateFormat)
		var filename string
		var title string
		if isFlagPassed("title") { // check if title flag is passed
			filename = formattedDate + " " + settings.Title + ".md"
			title = settings.Title
		} else {
			filename = formattedDate + ".md"
			title = ""
		}

		var description string
		if isFlagPassed("description") { // check if description flag is passed
			description = settings.Description
		} else {
			description = ""
		}

		data := struct {
			Title       string
			Description string
			CurrentDate string
		}{
			Title:       title,
			Description: description,
			CurrentDate: formattedDate,
		}

		var templatePath string
		if isFlagPassed("output-dir") { // check if output-dir flag is passed
			templatePath = path.Join(settings.OutputDirectory, filename)
		} else {
			templatePath = filename
		}

		// Create the template file in the output directory
		file, err := os.Create(templatePath)
		if err != nil {
			log.Fatalf("Error creating template file: %v", err)
		}
		defer file.Close()

		if err := tmpl.Execute(file, data); err != nil {
			log.Fatalf("Error executing template: %v", err)
		}

		fmt.Printf("Markdown template created at: %s\n", templatePath)
		return // Exit after creating the template
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

	if _, err := os.Stat(settings.InputDirectory); os.IsNotExist(err) {
		log.Fatalf("Input directory '%s' does not exist.", settings.InputDirectory)
	}
	if err := os.MkdirAll(settings.OutputDirectory, 0755); err != nil {
		log.Fatalf("Error creating output directory '%s': %v", settings.OutputDirectory, err)
	}

	switch *styleString {
	case "default":
		settings.Style = parse.Default
	case "dark":
		settings.Style = parse.Dark
	case "colorful":
		settings.Style = parse.Colorful
	default:
		settings.Style = parse.Default
	}

	// Initial build
	buildWebsite(settings)

	// Start watching for changes
	if *watch {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		// Add the input directory to the watcher
		err = watcher.Add(settings.InputDirectory)
		if err != nil {
			log.Fatal(err)
		}

		// Add custom css path, if any
		if settings.PathToCustomCss != "" {
			err = watcher.Add(settings.PathToCustomCss)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Add custom js path, if any
		if settings.PathToCustomJs != "" {
			err = watcher.Add(settings.PathToCustomJs)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Add custom favicon path, if any
		if settings.PathToCustomFavicon != "" {
			err = watcher.Add(settings.PathToCustomFavicon)
			if err != nil {
				log.Fatal(err)
			}
		}

		// serve files
		go serve(settings)

		log.Println("\nWatching for changes...\n")
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Changes detected. Rebuilding...")
					buildWebsite(settings)
					log.Println("\nWatching for changes...\n")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}
}

func serve(settings parse.Settings) {
	fmt.Printf("Serving '%s' on http://localhost:666\n", settings.OutputDirectory)
	http.Handle("/", http.FileServer(http.Dir(settings.OutputDirectory)))
	http.ListenAndServe(":666", nil)
}

func cleanContent(s string) []string {
	toReplace := []struct {
		old string
		new string
	}{
		{"’", "'"},
		{"–", " "},
	}

	toRemove := []string{
		"\n",
		"\r",
		"\t",
		"(",
		")",
		"[",
		"]",
		"{",
		"}",
		"\"",
		"\\",
		"/",
		"”",
		"#",
		"-",
		"*",
	}

	for _, pair := range toReplace {
		s = strings.ReplaceAll(s, pair.old, pair.new)
	}

	for _, char := range toRemove {
		s = strings.ReplaceAll(s, char, " ")
	}

	return strings.Fields(s)
}

func buildWebsite(settings parse.Settings) {
	// Clear output directory
	err := os.RemoveAll(settings.OutputDirectory)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(settings.OutputDirectory, 0755)
	if err != nil {
		log.Fatal(err)
	}

	files, err := parse.GetPaths(settings.InputDirectory, []string{".md", ".html"})
	if err != nil {
		log.Fatal(err)
	}

	var articles []parse.Article
	searchIndex := []map[string]interface{}{}
	for _, path := range files {
		article, err := processFile(path, settings)
		if err != nil {
			log.Printf("Error processing file %s: %s\n", path, err)
			continue
		}
		articles = append(articles, article)

		searchIndex = append(searchIndex, map[string]interface{}{
			"title":       article.Title,
			"content":     cleanContent(article.TextContent),
			"description": article.Description,
			"tags":        article.Tags,
			"url":         article.LinkToSelf,
		})
	}

	// Convert search index to JSON
	searchIndexJSON, err := json.Marshal(searchIndex)
	if err != nil {
		log.Fatal(err)
	}

	// Save search index JSON to a file
	searchIndexPath := filepath.Join(settings.OutputDirectory, "search_index.json")
	err = os.WriteFile(searchIndexPath, searchIndexJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

	err = parse.GenerateHtmlIndex(articles, settings)
	if err != nil {
		log.Fatal(err)
	}

	if settings.PathToCustomCss == "" {
		styleAsset := "style.css"
		switch settings.Style {
		case parse.Dark:
			styleAsset = "style-dark.css"
		case parse.Colorful:
			styleAsset = "style-colorful.css"
		}
		saveAsset(styleAsset, "style.css", settings.OutputDirectory)

	} else {
		input, err := os.ReadFile(settings.PathToCustomCss)
		if err != nil {
			log.Fatal(err)
		}

		cssDestPath := filepath.Join(settings.OutputDirectory, "style.css")
		err = os.WriteFile(cssDestPath, input, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	if settings.PathToCustomJs == "" {
		saveAsset("script.js", "script.js", settings.OutputDirectory)
	} else {
		input, err := os.ReadFile(settings.PathToCustomJs)
		if err != nil {
			panic(err)
		}
		jsDestPath := filepath.Join(settings.OutputDirectory, "script.js")
		err = os.WriteFile(jsDestPath, input, 0644)
		if err != nil {
			panic(err)
		}
	}

	if settings.PathToCustomFavicon == "" {
		saveAsset("favicon.ico", "favicon.ico", settings.OutputDirectory)
	} else {
		input, err := os.ReadFile(settings.PathToCustomFavicon)
		if err != nil {
			panic(err)
		}
		faviconDestPath := filepath.Join(settings.OutputDirectory, "favicon.ico")
		err = os.WriteFile(faviconDestPath, input, 0644)
		if err != nil {
			panic(err)
		}
	}

	saveAsset("search.js", "search.js", settings.OutputDirectory)

	log.Println("Blog generated successfully!")
}

func processFile(filePath string, settings parse.Settings) (parse.Article, error) {
	var article parse.Article
	var err error

	pathLower := strings.ToLower(filePath)

	if strings.HasSuffix(pathLower, ".md") {
		article, err = parse.MarkdownFile(filePath)
		parse.CopyHtmlResources(settings, &article)
		parse.FormatMarkdown(&article, settings)
	} else if strings.HasSuffix(filePath, ".html") {
		article, err = parse.HTMLFile(filePath)
		parse.CopyHtmlResources(settings, &article)
	} else {
		return parse.Article{}, fmt.Errorf("unsupported file type: %s", filePath)
	}
	if err != nil {
		panic(err)
	}

	os.WriteFile(article.LinkToSave, []byte(article.HtmlContent), 0644)
	return article, nil
}

func saveAsset(assetName string, saveName string, outputDirectory string) {
	file, err := assets.ReadFile("assets/" + assetName)
	if err != nil {
		log.Fatalf("Error reading asset '%s': %v", assetName, err)
	}

	pathToSave := path.Join(outputDirectory, saveName)
	if err := os.WriteFile(pathToSave, file, 0644); err != nil {
		log.Fatalf("Error saving asset '%s': %v", assetName, err)
	}
}
