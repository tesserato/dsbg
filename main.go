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
	"sort"
	"strings"
	texttemplate "text/template"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/tesserato/dsbg/parse"
)

//go:embed assets
var assets embed.FS

// isFlagPassed checks if a specific command-line flag was provided when running the program.
func isFlagPassed(name string) bool {
	nameToTest := strings.ToLower(name)
	nameToTest = strings.ReplaceAll(nameToTest, "-", "")
	nameToTest = strings.ReplaceAll(nameToTest, "_", "")
	found := false
	flag.Visit(func(f *flag.Flag) {
		nameFlag := strings.ToLower(f.Name)
		nameFlag = strings.ReplaceAll(nameFlag, "-", "")
		nameFlag = strings.ReplaceAll(nameFlag, "_", "")
		if nameFlag == nameToTest {
			found = true
		}
	})
	return found
}

// noFlagsPassed checks if any command-line flags were provided when running the program.
func noFlagsPassed(fs *flag.FlagSet) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		found = true
	})
	return !found
}

func logFlag(f *flag.Flag) {
	defaultValue := f.DefValue
	if defaultValue != "" {
		defaultValue = fmt.Sprintf(" (default: %v)", defaultValue)
	}
	fmt.Fprintf(os.Stderr, "  -%v %v\n    %v \n", f.Name, defaultValue, f.Usage)
}

func main() {
	// Create FlagSets for default flags and template flags
	defaultFlagSet := flag.NewFlagSet("default", flag.ExitOnError)
	templateFlagSet := flag.NewFlagSet("template", flag.ExitOnError)

	var settings parse.Settings

	// --- Default FlagSet Flags ---
	defaultFlagSet.StringVar(&settings.Title, "title", "Blog", "The title of the blog, used in the header and page titles.")
	defaultFlagSet.StringVar(&settings.BaseUrl, "base-url", "", "The base URL of the blog (e.g., https://example.com), used for generating absolute URLs in the RSS feed and sitemap.")
	defaultFlagSet.StringVar(&settings.DescriptionMarkdown, "description", "This is my blog", "A short description of the blog, used in meta tags and the index page. Supports Markdown formatting.")
	defaultFlagSet.StringVar(&settings.InputDirectory, "input-path", "content", "Path to the directory containing source content files (Markdown or HTML).")
	defaultFlagSet.StringVar(&settings.OutputDirectory, "output-path", "public", "Path to the directory where the generated website files will be saved.")
	defaultFlagSet.StringVar(&settings.DateFormat, "date-format", "2006 01 02", "Format for displaying dates on the website. Uses Go's time formatting (e.g., '2006-01-02' or 'January 2, 2006').")
	defaultFlagSet.StringVar(&settings.IndexName, "index-name", "index.html", "Filename for the main index page. Defaults to 'index.html'.")
	defaultFlagSet.StringVar(&settings.PathToCustomCss, "css-path", "", "Path to a custom CSS file to override the default styles. If not provided, a default style or a predefined style (using '-style') will be used.")
	defaultFlagSet.StringVar(&settings.PathToCustomJs, "js-path", "", "Path to a custom JavaScript file to include in the website. If not provided, default scripts will be used.")
	defaultFlagSet.StringVar(&settings.PathToCustomFavicon, "favicon-path", "", "Path to a custom favicon file (e.g., .ico, .png) to use for the website. If not provided, a default favicon will be used.")
	defaultFlagSet.BoolVar(&settings.DoNotExtractTagsFromPaths, "ignore-tags-from-paths", false, "Disable extracting tags from directory names. By default, directory names are used as tags.")
	defaultFlagSet.BoolVar(&settings.DoNotRemoveDateFromPaths, "keep-date-in-paths", false, "Do not remove date patterns (YYYY-MM-DD) from generated file paths. By default, dates are removed from paths for cleaner URLs.")
	defaultFlagSet.BoolVar(&settings.DoNotRemoveDateFromTitles, "keep-date-in-titles", false, "Do not remove date patterns (YYYY-MM-DD) from article titles. By default, dates are removed from titles.")
	defaultFlagSet.BoolVar(&settings.OpenInNewTab, "open-in-new-tab", false, "Open article links in a new browser tab using 'target=\"_blank\" rel=\"noopener\"'.")
	defaultFlagSet.StringVar(&settings.XHandle, "x-handle", "", "The handle to use for sharing on X.com (formerly Twitter). If provided, a share button will be included.")
	defaultFlagSet.StringVar(&settings.BlueSkyHandle, "bluesky-handle", "", "The handle to use for sharing on bsky.app (BlueSky). If provided, a share button will be included.")
	defaultFlagSet.StringVar(&settings.ThreadsHandle, "threads-handle", "", "The handle to use for sharing on threads.net (Threads). If provided, a share button will be included.")
	defaultFlagSet.StringVar(&settings.MastodonHandle, "mastodon-handle", "", "The handle to use for sharing on mastodon.social (Mastodon). If provided, a share button will be included.")
	defaultFlagSet.StringVar(&settings.Sort, "sort", "date-created", "How to sort articles on the index page. Options: date-created, reverse-date-created, date-updated, reverse-date-updated, title, reverse-title, path, reverse-path.")
	styleString := defaultFlagSet.String("style", "default", "Predefined style to use: 'default', 'dark', or 'colorful'. Overrides default styling but is overridden by a custom CSS file ('-css-path').")
	pathToAdditionalElementsTop := defaultFlagSet.String("elements-top", "", "Path to an HTML file with elements to include at the top of each page, inside the <head> tag (e.g., analytics scripts, custom meta tags).")
	pathToAdditionalElemensBottom := defaultFlagSet.String("elements-bottom", "", "Path to an HTML file with elements to include at the bottom of each page, before the closing </body> tag (e.g., additional scripts).")
	watch := defaultFlagSet.Bool("watch", false, "Watch for changes in the input directory and rebuild the website automatically. Also starts a local HTTP server to serve the generated website.")

	// --- Template FlagSet Flags ---
	var templateSettings parse.TemplateSettings
	templateFlagSet.StringVar(&templateSettings.Title, "title", "", "Title to populate the 'title' field.")
	templateFlagSet.StringVar(&templateSettings.Description, "description", "", "Description to populate the 'description' field.")
	templateFlagSet.StringVar(&templateSettings.Created, "created", "", "Date to populate the 'created' field. Must be in the format specified by the 'date-format' flag.")
	templateFlagSet.StringVar(&templateSettings.Updated, "updated", "", "Date to populate the 'updated' field. Must be in the format specified by the 'date-format' flag.")
	templateFlagSet.StringVar(&templateSettings.CoverImagePath, "cover-image-path", "", "Path to a cover image for the Markdown template file relative to the 'output-path' specified.")
	templateFlagSet.StringVar(&templateSettings.Tags, "tags", "", "A comma-separated list of tags for the Markdown template file.")
	templateFlagSet.StringVar(&templateSettings.OutputDirectory, "output-path", ".", "Path to the directory where the template will be saved.")
	templateFlagSet.StringVar(&settings.DateFormat, "date-format", "2006 01 02", "Format for displaying dates on the website. Uses Go's time formatting (e.g., '2006-01-02' or 'January 2, 2006').")

	defaultFlagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Custom help %s:\n", os.Args[0])

		defaultFlagSet.VisitAll(logFlag)

		fmt.Fprintf(os.Stderr, "\nMarkdown template flags:\n")

		templateFlagSet.VisitAll(logFlag)
	}

	switch os.Args[1] {
	case "template":
		log.Println("Creating Markdown template:")
		err := templateFlagSet.Parse(os.Args[1:])
		if err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}
		if err := createMarkdownTemplate(templateSettings); err != nil {
			log.Fatalf("Error creating markdown template: %v", err)
		}
		return

	}

	err := defaultFlagSet.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	// Convert Markdown description to HTML (Do this *after* parsing flags so description is populated)
	var buf strings.Builder
	if err := parse.Markdown.Convert([]byte(settings.DescriptionMarkdown), &buf); err != nil {
		log.Fatalf("failed to convert description to HTML: %v", err)
	}
	settings.DescriptionHTML = template.HTML(buf.String())

	// Check if the input directory exists
	if _, err := os.Stat(settings.InputDirectory); os.IsNotExist(err) {
		if noFlagsPassed(defaultFlagSet) {
			defaultFlagSet.Usage()
			return
		}
		log.Fatalf("Input directory '%s' does not exist.", settings.InputDirectory)
	}

	// Read content of files specified by flags
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

	// Ensure base URL is correctly formatted
	if settings.BaseUrl == "" {
		settings.BaseUrl = "http://localhost:666"
	} else {
		settings.BaseUrl = strings.TrimSuffix(settings.BaseUrl, "/")
	}

	// Create the output directory
	if err := os.MkdirAll(settings.OutputDirectory, 0755); err != nil {
		log.Fatalf("Error creating output directory '%s': %v", settings.OutputDirectory, err)
	}

	// Set the website style
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

	// Perform the initial website build
	buildWebsite(settings)

	if *watch {
		startWatcher(settings)
	}

}

// createMarkdownTemplate generates a Markdown template file with predefined frontmatter.
// This template is helpful for quickly creating new blog posts with consistent metadata structure.
func createMarkdownTemplate(templateSettings parse.TemplateSettings) error {
	tmpl, err := template.New("frontmatter").Parse(parse.FrontMatterTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	if templateSettings.Created == "" {
		templateSettings.Created = time.Now().Format(templateSettings.DateFormat)
	} else {
		parsed, err := time.Parse(templateSettings.DateFormat, templateSettings.Created)
		if err != nil {
			return fmt.Errorf("error parsing created date: %w", err)
		}
		templateSettings.Created = parsed.Format(templateSettings.DateFormat)
	}
	if templateSettings.Updated == "" {
		templateSettings.Updated = time.Now().Format(templateSettings.DateFormat)
	} else {
		parsed, err := time.Parse(templateSettings.DateFormat, templateSettings.Updated)
		if err != nil {
			return fmt.Errorf("error parsing updated date: %w", err)
		}
		templateSettings.Updated = parsed.Format(templateSettings.DateFormat)
	}
	formattedDate := time.Now().Format(templateSettings.DateFormat)
	filename := formattedDate + " " + templateSettings.Title + ".md"

	templatePath := filepath.Join(".", filename) // Default to current directory
	if isFlagPassed("output-path") {
		templatePath = filepath.Join(templateSettings.OutputDirectory, filename)
	}

	file, err := os.Create(templatePath)
	if err != nil {
		return fmt.Errorf("error creating template file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateSettings); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	fmt.Printf("Markdown template created at: %s\n", templatePath)
	return nil
}

// startWatcher initializes and starts the file system watcher.
// It monitors changes in the input directory and rebuilds the website automatically when changes are detected.
// It also starts a local HTTP server to serve the generated website for previewing.
func startWatcher(settings parse.Settings) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Add the input directory and custom asset paths to the watcher for monitoring changes.
	if err := watcher.Add(settings.InputDirectory); err != nil {
		log.Fatal(err)
	}

	// Recursively add all subdirectories within the input directory to the watcher.
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

	// Add custom asset paths to the watcher if they are specified.
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

	// Start the HTTP server in a goroutine to serve the website.
	go serve(settings)
	log.Printf("\n%s Watching for changes...\n", time.Now().Format(time.RFC850))
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Rebuild the website on write events, indicating file modifications.
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

// serve starts a simple HTTP server to serve the generated website from the output directory.
// It serves files on port 666 (or a configurable port, TODO: Allow the port to be specified).
func serve(settings parse.Settings) {
	addr := ":666" // TODO: Allow the port to be specified as a flag.
	fmt.Printf("Serving '%s' on http://localhost%s\n", settings.OutputDirectory, addr)
	http.Handle("/", http.FileServer(http.Dir(settings.OutputDirectory)))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// cleanContent prepares text content for search indexing by removing or replacing specific characters.
// This function aims to simplify the text for better search matching by removing punctuation and whitespace variations.
func cleanContent(s string) []string {
	replacements := map[string]string{
		"’": "'", // Replace curly apostrophe with straight apostrophe.
		"–": " ", // Replace en dash with space.
	}

	removals := []string{
		"\n", "\r", "\t", "(", ")", "[", "]", "{", "}",
		"\"", "\\", "/", "”", "#", "-", "*", // Remove various punctuation and formatting characters.
	}

	// Apply replacements.
	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}

	// Apply removals.
	for _, char := range removals {
		s = strings.ReplaceAll(s, char, " ")
	}

	// Split the string into words and return them as a slice.
	return strings.Fields(s)
}

// applyCSSTemplate loads the CSS template, parses it, and executes it with the selected theme data.
// This function is used when no custom CSS is provided and a predefined theme is selected.
func applyCSSTemplate(themeData parse.Theme, outputDirectory string) error {
	tmpl, err := texttemplate.ParseFS(assets, "assets/style-template.gocss")
	if err != nil {
		return fmt.Errorf("error parsing style template: %w", err)
	}

	var output strings.Builder
	// Execute the template with the theme data to generate the CSS content.
	err = tmpl.Execute(&output, themeData)
	if err != nil {
		return fmt.Errorf("error executing style template: %w", err)
	}

	pathToSave := filepath.Join(outputDirectory, "style.css")
	// Write the generated CSS content to the output file.
	if err := os.WriteFile(pathToSave, []byte(output.String()), 0644); err != nil {
		return fmt.Errorf("error saving processed css file: %w", err)
	}
	return nil
}

// getThemeData creates a theme struct with the appropriate settings based on the selected style.
// This function defines the visual theme of the website (colors, fonts) based on predefined styles.
func getThemeData(style parse.Style) parse.Theme {
	switch style {
	case parse.Dark:
		return parse.Theme{
			HeaderFont: "'Helvetica Neue', Helvetica, Arial, sans-serif",
			BodyFont:   "'Helvetica Neue', Helvetica, Arial, sans-serif",
			Background: "#303030",
			Text:       "#bfbfbf",
			Card:       "#3c3c3e",
			Link:       "#ff4b4b",
			Shadow:     "rgba(0, 0, 0, 0.777)",
		}
	case parse.Colorful:
		return parse.Theme{
			HeaderFont: "'Georgia', 'Times New Roman', Times, serif",
			BodyFont:   "'Raleway', sans-serif",
			Background: "#ffffff",
			Text:       "#000000",
			Card:       "#50d459a7",
			Button:     "#e65b5b",
			Link:       "#15598a",
			Shadow:     "rgba(98, 0, 0, 0.777)",
		}
	default: // Default style.
		return parse.Theme{
			HeaderFont: "\"Georgia\"",
			BodyFont:   "\"Garamond\"",
			Background: "#eaeaea",
			Text:       "#555555",
			Card:       "#ededed",
			Link:       "#c92626",
			Shadow:     "rgba(0, 0, 0, 0.25)",
		}
	}
}

// buildWebsite orchestrates the process of generating the website from the content files.
// This is the main function that controls the website generation process, including parsing files, generating index, RSS, and copying assets.
func buildWebsite(settings parse.Settings) {
	// Clear the output directory to ensure a clean build.
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

	var articles []parse.Article             // Slice to store processed articles.
	var searchIndex []map[string]interface{} // Slice to store data for the search index.

	// Process each content file.
	for _, path := range files {
		article, err := processFile(path, settings)
		if err != nil {
			log.Printf("Error processing file %s: %v\n", path, err)
			continue
		}
		articles = append(articles, article)

		// Add article data to the search index for client-side searching.
		searchIndex = append(searchIndex, map[string]interface{}{
			"title":       article.Title,
			"content":     cleanContent(article.TextContent), // Clean content for better search results.
			"description": article.Description,
			"tags":        article.Tags,
			"url":         article.LinkToSelf,
		})
	}

	// Sort articles based on the selected sorting method.
	switch settings.Sort {
	case "date-created":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Created.After(articles[j].Created)
		})
	case "reverse-date-created":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Created.Before(articles[j].Created)
		})
	case "date-updated":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Updated.After(articles[j].Updated)
		})
	case "reverse-date-updated":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Updated.Before(articles[j].Updated)
		})
	case "title":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Title < articles[j].Title
		})
	case "reverse-title":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Title > articles[j].Title
		})
	case "path":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].OriginalPath < articles[j].OriginalPath
		})
	case "reverse-path":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].OriginalPath > articles[j].OriginalPath
		})
	}

	// Generate and save the search index in JSON format.
	searchIndexJSON, err := json.Marshal(searchIndex)
	if err != nil {
		log.Fatalf("Error marshaling search index: %v", err)
	}
	searchIndexPath := filepath.Join(settings.OutputDirectory, "search_index.json")
	if err := os.WriteFile(searchIndexPath, searchIndexJSON, 0644); err != nil {
		log.Fatalf("Error saving search index: %v", err)
	}

	// Generate the HTML index page listing all articles.
	if err := parse.GenerateHtmlIndex(articles, settings, assets); err != nil {
		log.Fatalf("Error generating HTML index: %v", err)
	}

	// Generate the RSS feed for content syndication.
	if err := parse.GenerateRSS(articles, settings); err != nil {
		log.Fatalf("Error generating RSS feed: %v", err)
	}

	// Handle CSS: use custom CSS if provided, otherwise use predefined theme or default.
	if settings.PathToCustomCss == "" {
		theme := getThemeData(settings.Style)             // Get theme data based on selected style.
		applyCSSTemplate(theme, settings.OutputDirectory) // Apply CSS template with theme data.
	} else {
		// Copy custom CSS file to the output directory.
		if err := copyFile(settings.PathToCustomCss, filepath.Join(settings.OutputDirectory, "style.css")); err != nil {
			log.Fatalf("Error handling custom CSS: %v", err)
		}
	}

	// Handle JavaScript: use custom JS if provided, otherwise use default script.
	if settings.PathToCustomJs == "" {
		saveAsset("script.js", "script.js", settings.OutputDirectory) // Save default script.
	} else {
		// Copy custom JavaScript file to the output directory.
		if err := copyFile(settings.PathToCustomJs, filepath.Join(settings.OutputDirectory, "script.js")); err != nil {
			log.Fatalf("Error handling custom JavaScript: %v", err)
		}
	}

	// Handle favicon: use custom favicon if provided, otherwise use default favicon.
	if settings.PathToCustomFavicon == "" {
		saveAsset("favicon.ico", "favicon.ico", settings.OutputDirectory) // Save default favicon.
	} else {
		// Copy custom favicon file to the output directory.
		if err := copyFile(settings.PathToCustomFavicon, filepath.Join(settings.OutputDirectory, "favicon.ico")); err != nil {
			log.Fatalf("Error handling custom favicon: %v", err)
		}
	}

	// Save static assets (search script, social media icons).
	saveAsset("search.js", "search.js", settings.OutputDirectory)
	saveAsset("rss.svg", "rss.svg", settings.OutputDirectory)
	saveAsset("bluesky.svg", "bluesky.svg", settings.OutputDirectory)
	saveAsset("mastodon.svg", "mastodon.svg", settings.OutputDirectory)
	saveAsset("threads.svg", "threads.svg", settings.OutputDirectory)
	saveAsset("x.svg", "x.svg", settings.OutputDirectory)
	saveAsset(("share.svg"), "share.svg", settings.OutputDirectory)
	saveAsset("follow.svg", "follow.svg", settings.OutputDirectory)

	log.Println("Blog generated successfully!")
}

// processFile reads and parses a single content file (Markdown or HTML).
// It determines the file type, parses it accordingly, and prepares it for website generation.
func processFile(filePath string, settings parse.Settings) (parse.Article, error) {
	var article parse.Article
	var err error

	filePathLower := strings.ToLower(filePath) // Convert file path to lowercase for case-insensitive suffix check.

	// Process Markdown files.
	if strings.HasSuffix(filePathLower, ".md") {
		article, err = parse.MarkdownFile(filePath) // Parse Markdown file.
		if err != nil {
			return parse.Article{}, fmt.Errorf("error parsing markdown file: %w", err)
		}
		if err := parse.CopyHtmlResources(settings, &article); err != nil { // Copy any HTML resources linked in Markdown.
			return parse.Article{}, fmt.Errorf("error copying resources for markdown file: %w", err)
		}
		if err := parse.FormatMarkdown(&article, settings, assets); err != nil { // Format Markdown content (e.g., extract metadata, generate paths).
			return parse.Article{}, fmt.Errorf("error formatting markdown: %w", err)
		}
	} else if strings.HasSuffix(filePath, ".html") { // Process HTML files.
		article, err = parse.HTMLFile(filePath) // Parse HTML file.
		if err != nil {
			return parse.Article{}, fmt.Errorf("error parsing HTML file: %w", err)
		}
		if err := parse.CopyHtmlResources(settings, &article); err != nil { // Copy any HTML resources linked in HTML.
			return parse.Article{}, fmt.Errorf("error copying resources for HTML file: %w", err)
		}
	} else {
		return parse.Article{}, fmt.Errorf("unsupported file type: %s", filePath)
	}

	// Write the processed HTML content of the article to the output directory.
	if err := os.WriteFile(article.LinkToSave, []byte(article.HtmlContent), 0644); err != nil {
		return parse.Article{}, fmt.Errorf("error writing processed file: %w", err)
	}
	return article, nil
}

// saveAsset reads an asset file from the embedded filesystem and saves it to the output directory.
// This function is used to deploy default assets like scripts, styles, and icons that are bundled within the application.
func saveAsset(assetName string, saveName string, outputDirectory string) {
	file, err := assets.ReadFile("assets/" + assetName) // Read asset from embedded filesystem.
	if err != nil {
		log.Fatalf("Error reading asset '%s': %v", assetName, err)
	}

	pathToSave := filepath.Join(outputDirectory, saveName)
	// Write the asset file to the output directory.
	if err := os.WriteFile(pathToSave, file, 0644); err != nil {
		log.Fatalf("Error saving asset '%s': %v", assetName, err)
	}
}

// copyFile copies a file from a source path to a destination path.
// This utility function is used for copying custom CSS, JavaScript, and favicon files to the output directory.
func copyFile(srcPath string, destPath string) error {
	input, err := os.ReadFile(srcPath) // Read the source file.
	if err != nil {
		return fmt.Errorf("error reading file '%s': %w", srcPath, err)
	}
	err = os.WriteFile(destPath, input, 0644) // Write the content to the destination file.
	if err != nil {
		return fmt.Errorf("error writing file '%s': %w", destPath, err)
	}
	return nil
}
