package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"dsbg/parse"
)

//go:embed assets
var assets embed.FS

// isFlagPassed checks if a specific flag was provided when running the program.
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

	// Define command-line flags with descriptions.
	flag.StringVar(&settings.Title, "title", "Blog", "The title of the blog")
	flag.StringVar(&settings.BaseUrl, "base-url", "", "The base URL of the blog (e.g., https://example.com)")
	flag.StringVar(&settings.Description, "description", "This is my blog", "A short description of the blog")
	flag.StringVar(&settings.InputDirectory, "input-path", "content", "Path to the directory containing source files (Markdown or HTML)")
	flag.StringVar(&settings.OutputDirectory, "output-path", "public", "Path to the directory where generated website files will be saved")
	flag.StringVar(&settings.DateFormat, "date-format", "2006 01 02", "Format for displaying dates")
	flag.StringVar(&settings.IndexName, "index-name", "index.html", "Filename for the main index page")
	flag.StringVar(&settings.PathToCustomCss, "css-path", "", "Path to a custom CSS file")
	flag.StringVar(&settings.PathToCustomJs, "js-path", "", "Path to a custom JavaScript file")
	flag.StringVar(&settings.PathToCustomFavicon, "favicon-path", "", "Path to a custom favicon file (e.g., .ico)")
	flag.BoolVar(&settings.DoNotExtractTagsFromPaths, "ignore-tags-from-paths", false, "Disable extracting tags from directory names")
	flag.BoolVar(&settings.DoNotRemoveDateFromPaths, "keep-date-in-paths", false, "Do not remove date patterns from generated file paths")
	flag.BoolVar(&settings.DoNotRemoveDateFromTitles, "keep-date-in-titles", false, "Do not remove date patterns from article titles")
	flag.BoolVar(&settings.OpenInNewTab, "open-in-new-tab", false, "Open article links in a new browser tab")
	styleString := flag.String("style", "default", "Predefined style to use (default, dark, colorful)")
	pathToAdditionalElementsTop := flag.String("elements-top", "", "Path to an HTML file with elements to include at the top of each page (e.g., analytics scripts)")
	pathToAdditionalElemensBottom := flag.String("elements-bottom", "", "Path to an HTML file with elements to include at the bottom of each page")
	showHelp := flag.Bool("help", false, "Show this help message and exit")
	watch := flag.Bool("watch", false, "Watch for changes in the input directory and rebuild automatically. Also creates a server to serve the website.")
	createTemplate := flag.Bool("template", false, "Create a basic Markdown template file with frontmatter")

	flag.Parse()

    // Display help message and exit if the "help" flag is enabled.
	if *showHelp {
		flag.Usage()
		return
	}

	// Create a basic Markdown template and exit if the "template" flag is enabled.
	if *createTemplate {
		if err := createMarkdownTemplate(settings); err != nil {
			log.Fatalf("Error creating markdown template: %v", err)
		}
		return
	}

	// Read content of files specified by flags and store them in settings.
	if *pathToAdditionalElementsTop != "" {
		content, err := os.ReadFile(*pathToAdditionalElementsTop)
		if err != nil {
			log.Fatalf("Error reading additional top elements file: %v", err)
		}
		settings.AdditionalElementsTop = template.HTML(content)
	}

	if *pathToAdditionalElemensBottom != "" {
		content, err := os.ReadFile(*pathToAdditionalElemensBottom)
		if err != nil {
			log.Fatalf("Error reading additional bottom elements file: %v", err)
		}
		settings.AdditionalElemensBottom = template.HTML(content)
	}

	// Check if the input directory exists.
	if _, err := os.Stat(settings.InputDirectory); os.IsNotExist(err) {
		log.Fatalf("Input directory '%s' does not exist.", settings.InputDirectory)
	}

	// Create the output directory if it doesn't exist.
	if err := os.MkdirAll(settings.OutputDirectory, 0755); err != nil {
		log.Fatalf("Error creating output directory '%s': %v", settings.OutputDirectory, err)
	}

	// Set the website style based on the provided flag.
	switch *styleString {
	case "default":
		settings.Style = parse.Default
	case "dark":
		settings.Style = parse.Dark
	case "colorful":
		settings.Style = parse.Colorful
	default:
		settings.Style = parse.Default
		log.Printf("Unknown style '%s', using default.\n", *styleString)
	}

	// Perform the initial website build.
	buildWebsite(settings)

	// Start the watcher if the "watch" flag is enabled.
	if *watch {
		startWatcher(settings)
	}
}

// createMarkdownTemplate generates a Markdown template file with predefined frontmatter.
func createMarkdownTemplate(settings parse.Settings) error {
	tmpl, err := template.New("frontmatter").Parse(parse.FrontMatterTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	formattedDate := time.Now().Format(settings.DateFormat)
	filename := formattedDate + " " + settings.Title + ".md"
	if !isFlagPassed("title") {
		filename = formattedDate + ".md"
		settings.Title = "" // Ensure title is empty if flag not passed
	}

	description := settings.Description
	if !isFlagPassed("description") {
		description = "" // Ensure description is empty if flag not passed
	}

	data := struct {
		Title       string
		Description string
		CurrentDate string
	}{
		Title:       settings.Title,
		Description: description,
		CurrentDate: formattedDate,
	}

	templatePath := filepath.Join(".", filename) // Default to current directory
	if isFlagPassed("output-dir") {
		templatePath = filepath.Join(settings.OutputDirectory, filename)
	}

	file, err := os.Create(templatePath)
	if err != nil {
		return fmt.Errorf("error creating template file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	fmt.Printf("Markdown template created at: %s\n", templatePath)
	return nil
}

// startWatcher initializes and starts the file system watcher.
func startWatcher(settings parse.Settings) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Add the input directory and custom asset paths to the watcher.
    if err := watcher.Add(settings.InputDirectory); err != nil {
		log.Fatal(err)
	}
    
    err = filepath.WalkDir(settings.InputDirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
            if err := watcher.Add(path); err != nil {
                log.Fatal(err)
            }
		}
		return nil
	})

    if err != nil {
        log.Fatal(err)
    }

	if settings.PathToCustomCss != "" {
		if err := watcher.Add(settings.PathToCustomCss); err != nil {
			log.Fatal(err)
		}
	}
	if settings.PathToCustomJs != "" {
		if err := watcher.Add(settings.PathToCustomJs); err != nil {
			log.Fatal(err)
		}
	}
	if settings.PathToCustomFavicon != "" {
		if err := watcher.Add(settings.PathToCustomFavicon); err != nil {
			log.Fatal(err)
		}
	}

	// Start the file server.
    // time.Sleep(2 * time.Second)
	go serve(settings)
	log.Printf("\n%s Watching for changes...\n", time.Now().Format(time.RFC850))
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Rebuild the website on write events.
			if event.Has(fsnotify.Write) {
				log.Println("Changes detected. Rebuilding...")
				buildWebsite(settings)
                log.Printf("\n%s Watching for changes...\n", time.Now().Format(time.RFC850))
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

// serve starts a simple HTTP server to serve the generated website.
func serve(settings parse.Settings) {
	addr := ":666" // TODO: Allow the port to be specified.
	fmt.Printf("Serving '%s' on http://localhost%s\n", settings.OutputDirectory, addr)
	http.Handle("/", http.FileServer(http.Dir(settings.OutputDirectory)))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// cleanContent prepares text content for search indexing by removing or replacing specific characters.
func cleanContent(s string) []string {
	replacements := map[string]string{
		"’": "'",
		"–": " ",
	}

	removals := []string{
		"\n", "\r", "\t", "(", ")", "[", "]", "{", "}",
		"\"", "\\", "/", "”", "#", "-", "*",
	}

	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}

	for _, char := range removals {
		s = strings.ReplaceAll(s, char, " ")
	}

	return strings.Fields(s)
}

// buildWebsite orchestrates the process of generating the website from the content files.
func buildWebsite(settings parse.Settings) {
	// Clear the output directory.
	if err := os.RemoveAll(settings.OutputDirectory); err != nil {
		log.Fatalf("Error clearing output directory: %v", err)
	}
	if err := os.MkdirAll(settings.OutputDirectory, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Find all Markdown and HTML files in the input directory.
	files, err := parse.GetPaths(settings.InputDirectory, []string{".md", ".html"})
	if err != nil {
		log.Fatalf("Error getting content files: %v", err)
	}

	var articles []parse.Article
	var searchIndex []map[string]interface{}

	// Process each content file.
	for _, path := range files {
		article, err := processFile(path, settings)
		if err != nil {
			log.Printf("Error processing file %s: %v\n", path, err)
			continue
		}
		articles = append(articles, article)

		// Add article data to the search index.
		searchIndex = append(searchIndex, map[string]interface{}{
			"title":       article.Title,
			"content":     cleanContent(article.TextContent),
			"description": article.Description,
			"tags":        article.Tags,
			"url":         article.LinkToSelf,
		})
	}

	// Generate and save the search index.
	searchIndexJSON, err := json.Marshal(searchIndex)
	if err != nil {
		log.Fatalf("Error marshaling search index: %v", err)
	}
	searchIndexPath := filepath.Join(settings.OutputDirectory, "search_index.json")
	if err := os.WriteFile(searchIndexPath, searchIndexJSON, 0644); err != nil {
		log.Fatalf("Error saving search index: %v", err)
	}

	// Generate the HTML index page.
	if err := parse.GenerateHtmlIndex(articles, settings); err != nil {
		log.Fatalf("Error generating HTML index: %v", err)
	}

	// Generate the RSS feed.
	if err := parse.GenerateRSS(articles, settings); err != nil {
		log.Fatalf("Error generating RSS feed: %v", err)
	}

	// Handle CSS (custom or default).
	if settings.PathToCustomCss == "" {
		var styleAsset string
		switch settings.Style {
		case parse.Dark:
			styleAsset = "style-dark.css"
		case parse.Colorful:
			styleAsset = "style-colorful.css"
		default:
			styleAsset = "style.css"
		}
		saveAsset(styleAsset, "style.css", settings.OutputDirectory)
	} else {
		if err := copyFile(settings.PathToCustomCss, filepath.Join(settings.OutputDirectory, "style.css")); err != nil {
			log.Fatalf("Error handling custom CSS: %v", err)
		}
	}

	// Handle JavaScript (custom or default).
	if settings.PathToCustomJs == "" {
		saveAsset("script.js", "script.js", settings.OutputDirectory)
	} else {
		if err := copyFile(settings.PathToCustomJs, filepath.Join(settings.OutputDirectory, "script.js")); err != nil {
			log.Fatalf("Error handling custom JavaScript: %v", err)
		}
	}

	// Handle favicon (custom or default).
	if settings.PathToCustomFavicon == "" {
		saveAsset("favicon.ico", "favicon.ico", settings.OutputDirectory)
	} else {
		if err := copyFile(settings.PathToCustomFavicon, filepath.Join(settings.OutputDirectory, "favicon.ico")); err != nil {
			log.Fatalf("Error handling custom favicon: %v", err)
		}
	}

	// Save static assets.
	saveAsset("search.js", "search.js", settings.OutputDirectory)

	log.Println("Blog generated successfully!")
}

// processFile reads and parses a single content file (Markdown or HTML).
func processFile(filePath string, settings parse.Settings) (parse.Article, error) {
	var article parse.Article
	var err error

	filePathLower := strings.ToLower(filePath)

	if strings.HasSuffix(filePathLower, ".md") {
		article, err = parse.MarkdownFile(filePath)
		if err != nil {
			return parse.Article{}, fmt.Errorf("error parsing markdown file: %w", err)
		}
		if err := parse.CopyHtmlResources(settings, &article); err != nil {
			return parse.Article{}, fmt.Errorf("error copying resources for markdown file: %w", err)
		}
		if err := parse.FormatMarkdown(&article, settings); err != nil {
			return parse.Article{}, fmt.Errorf("error formatting markdown: %w", err)
		}
	} else if strings.HasSuffix(filePath, ".html") {
		article, err = parse.HTMLFile(filePath)
		if err != nil {
			return parse.Article{}, fmt.Errorf("error parsing HTML file: %w", err)
		}
		if err := parse.CopyHtmlResources(settings, &article); err != nil {
			return parse.Article{}, fmt.Errorf("error copying resources for HTML file: %w", err)
		}
	} else {
		return parse.Article{}, fmt.Errorf("unsupported file type: %s", filePath)
	}

	if err := os.WriteFile(article.LinkToSave, []byte(article.HtmlContent), 0644); err != nil {
		return parse.Article{}, fmt.Errorf("error writing processed file: %w", err)
	}
	return article, nil
}

// saveAsset reads an asset file from the embedded filesystem and saves it to the output directory.
func saveAsset(assetName string, saveName string, outputDirectory string) {
	file, err := assets.ReadFile("assets/" + assetName)
	if err != nil {
		log.Fatalf("Error reading asset '%s': %v", assetName, err)
	}

	pathToSave := filepath.Join(outputDirectory, saveName)
	if err := os.WriteFile(pathToSave, file, 0644); err != nil {
		log.Fatalf("Error saving asset '%s': %v", assetName, err)
	}
}

// copyFile copies a file from a source path to a destination path.
func copyFile(srcPath string, destPath string) error {
	input, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("error reading file '%s': %w", srcPath, err)
	}
	err = os.WriteFile(destPath, input, 0644)
	if err != nil {
		return fmt.Errorf("error writing file '%s': %w", destPath, err)
	}
	return nil
}