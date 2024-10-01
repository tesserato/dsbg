package main

import (
	"dsbg/parse"
	"embed"
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
	flag.StringVar(&settings.InputDirectory, "input-dir", "content", "Path to the directory that holds the source files")
	flag.StringVar(&settings.OutputDirectory, "output-dir", "public", "Path to the directory where the output files will be saved")
	flag.StringVar(&settings.DateFormat, "date-format", "2006 01 02", "Date format")
	flag.StringVar(&settings.IndexName, "index-name", "index.html", "Name of the index files")
	flag.StringVar(&settings.PathToCustomCss, "path-to-custom-css", "", "Path to a file with custom css")
	flag.StringVar(&settings.PathToCustomJs, "path-to-custom-js", "", "Path to a file with custom js")
	styleString := flag.String("style", "default", "Style to be used")
	pathToAdditionalElementsTop := flag.String("path-to-additional-elements-top", "", "Path to a file with additional elements (basically scripts) to be placed at the top of the HTML outputs")
	pathToAdditionalElemensBottom := flag.String("path-to-additional-elements-bottom", "", "Path to a file with additional elements (basically scripts) to be placed at the bottom of the HTML outputs")
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
		// Data for the template

		// Get filename from title or default to "template.md"
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
	http.Handle("/", http.FileServer(http.Dir(settings.OutputDirectory)))
	http.ListenAndServe(":666", nil)
}

func buildWebsite(settings parse.Settings) {
	files, err := parse.GetPaths(settings.InputDirectory, []string{".md", ".html"})
	if err != nil {
		log.Fatal(err)
	}

	var articles []parse.Article
	for _, path := range files {
		article, err := processFile(path, settings)
		if err != nil {
			log.Printf("Error processing file %s: %s\n", path, err)
			continue
		}
		articles = append(articles, article)
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
		}
		saveAsset(styleAsset, "style.css", settings.OutputDirectory)

	} else {
		input, err := os.ReadFile(settings.PathToCustomCss)
		if err != nil {
			panic(err)
		}

		cssDestPath := filepath.Join(settings.OutputDirectory, "style.css")
		err = os.WriteFile(cssDestPath, input, 0644)
		if err != nil {
			panic(err)
		}
	}

	if settings.PathToCustomJs == "" {
		saveAsset("script.js", "script.js", settings.OutputDirectory)
	} else{
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

	saveAsset("favicon.ico", "favicon.ico", settings.OutputDirectory)

	log.Println("Blog generated successfully!")
}
func processFile(filePath string, settings parse.Settings) (parse.Article, error) {
	var article parse.Article
	var err error
	var links parse.Links

	pathLower := strings.ToLower(filePath)

	if strings.HasSuffix(pathLower, ".md") {
		article, err = parse.MarkdownFile(filePath)
		links = parse.CopyHtmlResources(settings, filePath, article.HtmlContent)
		article = parse.FormatMarkdown(article, links, settings)
	} else if strings.HasSuffix(filePath, ".html") {
		article, err = parse.HTMLFile(filePath)
		links = parse.CopyHtmlResources(settings, filePath, article.HtmlContent)
	} else {
		return parse.Article{}, fmt.Errorf("unsupported file type: %s", filePath)
	}

	if err != nil {
		return parse.Article{}, err
	}

	os.WriteFile(links.ToSave, []byte(article.HtmlContent), 0644)
	article = parse.FormatMarkdown(article, links, settings)

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
