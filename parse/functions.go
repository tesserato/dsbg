package parse

import (
	"bytes"
	"fmt"
	"github.com/k3a/html2text"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
	"golang.org/x/net/html"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	texttemplate "text/template"
	"time"
)

// regexPatterns defines a list of regular expression patterns to identify dates in strings.
var regexPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?P<year>\d{4})\D+(?P<month>\d{1,2})\D+(?P<day>\d{1,2})`),
	regexp.MustCompile(`(?P<day>\d{1,2})\D+(?P<month>\d{1,2})\D+(?P<year>\d{4})`),
	regexp.MustCompile(`(?P<hour>\d{2}):(?P<min>\d{2}):(?P<sec>\d{2})`),
}

// RemoveDateFromPath attempts to remove date patterns from a given string.
// It iterates through predefined regex patterns and replaces matching substrings with an empty string.
// Finally, it trims any leading/trailing hyphens, underscores, or spaces.
func RemoveDateFromPath(stringWithDate string) string {
	for _, regexPattern := range regexPatterns {
		stringWithDate = regexPattern.ReplaceAllString(stringWithDate, "")
	}
	stringWithDate = strings.Trim(stringWithDate, "-_ ")
	return stringWithDate
}

// DateTimeFromString attempts to parse a date and time from a string using predefined regex patterns.
// It extracts named capture groups (year, month, day, hour, min, sec) and constructs a time.Time value.
// It panics if a matched group cannot be converted to an integer.
func DateTimeFromString(date string) time.Time {
	m := make(map[string]int)
	for _, pattern := range regexPatterns {
		matches := pattern.FindStringSubmatch(date)
		if len(matches) > 0 {
			for i, name := range pattern.SubexpNames()[1:] {
				integer, err := strconv.Atoi(matches[i+1])
				if err != nil {
					panic(err) // Potential Problem: Panicking here will crash the application. Consider returning an error.
				}
				m[name] = integer
			}
		}
	}
	year := m["year"]
	month := time.Month(m["month"])
	day := m["day"]
	hour := m["hour"]
	min := m["min"]
	sec := m["sec"]
	dateTime := time.Date(year, month, day, hour, min, sec, 0, time.UTC)
	return dateTime
}

// GetPaths retrieves all file paths within a directory and its subdirectories
// that match the specified extensions.
func GetPaths(root string, extensions []string) ([]string, error) {
	var files []string
	extMap := make(map[string]bool) // Create a map for efficient extension lookup

	for _, ext := range extensions {
		extLower := strings.ToLower(strings.TrimSpace(ext)) // Normalize extension (lowercase, trim whitespace)
		extMap[extLower] = true                             // Add normalized extension to the map for faster checking
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(path)) // Get and normalize file extension
			if extMap[ext] {                           // Check if extension is in the allowed set
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

// cleanString removes non-alphanumeric characters (except '/', '\', '.'), replaces backslashes with forward slashes,
// trims hyphens, underscores, and spaces from path segments, and joins them with hyphens.
func cleanString(url string) string {
	var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9\/\\\. ]+`)
	url = nonAlphanumericRegex.ReplaceAllString(url, "")
	url = strings.ReplaceAll(url, "\\", "/")
	pieces := strings.Split(url, "/")
	for i, piece := range pieces {
		pieces[i] = strings.Trim(piece, "-_ ")
	}
	url = strings.Join(pieces, "/")
	pieces = strings.Fields(url)
	for i, piece := range pieces {
		pieces[i] = strings.Trim(piece, "-_ ")
	}
	url = strings.Join(pieces, "-")

	url = strings.Trim(url, "-")
	return url
}

// CopyHtmlResources copies associated resources (like images, scripts) for an article,
// determines the output path, and handles special cases for "PAGE" tagged articles.
func CopyHtmlResources(settings Settings, article *Article) {
	relativeInputPath, err := filepath.Rel(settings.InputDirectory, article.OriginalPath)
	if err != nil {
		panic(err) // Potential Problem: Panicking here will crash the application. Consider returning an error.
	}

	// Optionally remove date from the article title based on settings.
	if !settings.DoNotRemoveDateFromTitles {
		datelessTitle := RemoveDateFromPath(article.Title)
		if datelessTitle != "" {
			article.Title = datelessTitle
		}
	}

	// Optionally extract tags from the input path based on settings.
	if !settings.DoNotExtractTagsFromPaths {
		relativeInputPathNoDate := RemoveDateFromPath(relativeInputPath)
		relativeInputPathNoDate = filepath.Clean(relativeInputPathNoDate)
		pathTags := strings.Split(relativeInputPathNoDate, string(os.PathSeparator))
		for i, tag := range pathTags {
			pathTags[i] = strings.Trim(tag, "-_ ")
		}
		if len(pathTags) > 1 {
			pathTags = pathTags[:len(pathTags)-1]
			article.Tags = append(article.Tags, pathTags...)
		}
	}

	// Determine the output path for the article.
	outputPath := filepath.Join(settings.OutputDirectory, relativeInputPath)
	outputPath = strings.TrimSuffix(outputPath, filepath.Ext(outputPath))
	outputPath = filepath.Join(outputPath, settings.IndexName)

	// Optionally remove date from the output path based on settings.
	if !settings.DoNotRemoveDateFromPaths {
		datelessOutputPath := RemoveDateFromPath(outputPath)
		if !(strings.Contains(datelessOutputPath, "\\") || strings.Contains(datelessOutputPath, "//")) {
			outputPath = datelessOutputPath
		}
	}
	outputPath = cleanString(outputPath)
	outputDirectory := filepath.Dir(outputPath)
	err = os.MkdirAll(outputDirectory, os.ModePerm)
	if err != nil {
		panic(err) // Potential Problem: Panicking here will crash the application. Consider returning an error.
	}

	originalDirectory := filepath.Dir(article.OriginalPath)

	// Special handling for articles tagged as "PAGE". Copies the entire original directory.
	if slices.Contains(article.Tags, "PAGE") && originalDirectory != settings.InputDirectory {
		visit := func(originalPath string, di fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !di.Type().IsRegular() { // Skip non-regular files (e.g., directories, symlinks, devices)
				switch di.Type() {
				case fs.ModeSymlink:
					fmt.Printf("Skipping symlink: %s\n", originalPath)
				case fs.ModeDevice:
					fmt.Printf("Skipping device: %s\n", originalPath)
				case fs.ModeNamedPipe:
					fmt.Printf("Skipping named pipe: %s\n", originalPath)
				case fs.ModeSocket:
					fmt.Printf("Skipping socket: %s\n", originalPath)
				case fs.ModeDir:
					return nil
				default:
					fmt.Printf("Skipping non-regular file: %s\n", originalPath)
				}
				return nil // Skip, but don't consider it an error
			}

			relativeOriginalPath, err := filepath.Rel(originalDirectory, originalPath)
			if err != nil {
				return fmt.Errorf("error getting relative path for %s: %w", originalPath, err) // Wrap error for better context
			}

			destPath := filepath.Join(outputDirectory, relativeOriginalPath)
			destFolder := filepath.Dir(destPath)
			err = os.MkdirAll(filepath.FromSlash(destFolder), 0755)
			if err != nil {
				return fmt.Errorf("error creating directories for %s: %w", destPath, err) // Wrap error
			}

			file, err := os.ReadFile(originalPath)
			if err != nil {
				return fmt.Errorf("error reading file %s: %w", originalPath, err) // Wrap error
			}

			err = os.WriteFile(destPath, file, 0644)
			if err != nil {
				return fmt.Errorf("error writing file %s: %w", destPath, err) // Wrap error
			}

			fmt.Printf("%s -> %s\n", originalPath, destPath)
			return nil
		}

		err = filepath.WalkDir(originalDirectory, visit)
		if err != nil {
			panic(err) // Potential Problem: Panicking here will crash the application. Consider returning an error.
		}
	}

	// Copy individual resources (images, scripts, etc.) linked in the HTML content.
	for _, resourceOrigRelPath := range extractResources(article.HtmlContent) {
		resourceOrigRelPathLower := strings.ToLower(resourceOrigRelPath)
		if strings.Contains(resourceOrigRelPathLower, "http") {
			continue
		}
		resourceOrigPath := filepath.Join(originalDirectory, resourceOrigRelPath)
		resourceDestPath := filepath.Join(outputDirectory, resourceOrigRelPath)
		fmt.Printf("  Copying %s to %s\n", resourceOrigPath, resourceDestPath)

		input, err := os.ReadFile(resourceOrigPath)
		if err != nil {
			panic(err) // Potential Problem: Panicking here will crash the application. Consider returning an error.
		}

		err = os.MkdirAll(filepath.Dir(filepath.FromSlash(resourceDestPath)), 0755)
		if err != nil {
			panic(err) // Potential Problem: Panicking here will crash the application. Consider returning an error.
		}

		err = os.WriteFile(resourceDestPath, input, 0644)
		if err != nil {
			panic(err) // Potential Problem: Panicking here will crash the application. Consider returning an error.
		}
	}

	// Set the relative link to the generated article.
	LinkToSelf, err := filepath.Rel(settings.OutputDirectory, outputPath)
	if err != nil {
		panic(err) // Potential Problem: Panicking here will crash the application. Consider returning an error.
	}
	article.LinkToSelf = filepath.ToSlash(LinkToSelf)
	article.LinkToSave = filepath.ToSlash(outputPath)
	// fmt.Printf(
	// 	"InputDirectory: %s\noriginalArticlePath: %s\nrelativeInputPath: %s\noutputDirectory: %s\noutputPath: %s\nLinkToSelf: %s\n\n",
	// 	settings.InputDirectory, article.OriginalPath, relativeInputPath, outputDirectory, article.LinkToSave, article.LinkToSelf)
}

// GenerateHtmlIndex creates an HTML index page listing all processed articles.
func GenerateHtmlIndex(articles []Article, settings Settings) error {
	// Separate articles into pages and regular articles based on tags.
	var allTags []string
	var pageList []Article
	var articleList []Article
	for _, article := range articles {
		if slices.Contains(article.Tags, "PAGE") {
			pageList = append(pageList, article)
		} else {
			allTags = append(allTags, article.Tags...)
			articleList = append(articleList, article)
		}
	}

	// Define template functions.
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

	// Execute the template with article data.
	var tp bytes.Buffer
	err = tmpl.Execute(&tp, struct {
		AllTags     []string
		PageList    []Article
		ArticleList []Article
		Settings    Settings
	}{allTags, pageList, articleList, settings})
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	// Write the generated HTML to the output file.
	filePath := filepath.Join(settings.OutputDirectory, settings.IndexName)
	return os.WriteFile(filePath, tp.Bytes(), 0644)
}

// GenerateRSS creates an RSS feed XML file from the processed articles.
func GenerateRSS(articles []Article, settings Settings) error {
	// Parse the RSS template.
	tmpl, err := texttemplate.New("rss.xml").Parse(rssTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	// Execute the template with article data.
	var tp bytes.Buffer
	err = tmpl.Execute(&tp, struct {
		Articles []Article
		Settings Settings
	}{articles, settings})
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	// Write the generated RSS feed to the output file.
	filePath := filepath.Join(settings.OutputDirectory, "rss.xml")
	return os.WriteFile(filePath, tp.Bytes(), 0644)
}

// MarkdownFile parses a Markdown file, extracts frontmatter, and populates an Article struct.
func MarkdownFile(path string) (Article, error) {
	// Read the Markdown file content.
	data, err := os.ReadFile(path)
	if err != nil {
		return Article{}, err
	}

	// Configure Goldmark Markdown parser with frontmatter support.
	markdown := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAttribute(),
			parser.WithAutoHeadingID(),
		),
		goldmark.WithExtensions(
			&frontmatter.Extender{},
		),
	)

	// Create a context to store frontmatter.
	context := parser.NewContext()

	// Parse the Markdown content and render to HTML, storing frontmatter in the context.
	var buf strings.Builder
	if err := markdown.Convert(data, &buf, parser.WithContext(context)); err != nil {
		panic(err) // Potential Problem: Panicking here will crash the application. Consider returning an error.
	}
	// content := buf.String()

	// Retrieve frontmatter from the context.
	var article = Article{OriginalPath: path, TextContent: string(data), HtmlContent: buf.String()}
	fm := frontmatter.Get(context)
	if fm != nil {
		var d map[string]any
		if err := fm.Decode(&d); err != nil {
			fmt.Printf("Article path: %s", path)
			panic(err) // Potential Problem: Panicking here will crash the application. Consider returning an error.
		}
		// Populate Article fields from frontmatter.
		for name, value := range d {
			name = strings.ToLower(name)
			name = strings.Trim(name, " ")
			if value == nil {
				continue
			}
			switch name {
			case "title":
				article.Title = value.(string)
			case "description":
				article.Description = value.(string)
			case "created":
				if reflect.TypeOf(value).Kind() == reflect.String {
					article.Created = DateTimeFromString(value.(string))
				} else {
					article.Created = value.(time.Time)
				}
			case "updated":
				if reflect.TypeOf(value).Kind() == reflect.String {
					article.Updated = DateTimeFromString(value.(string))
				} else {
					article.Updated = value.(time.Time)
				}
			case "tags":
				switch reflect.TypeOf(value).Kind() {

				case reflect.Slice:
					tags := value.([]any)
					for _, tag := range tags {
						tagString := strings.Trim(tag.(string), " ")
						article.Tags = append(article.Tags, tagString)
					}

				case reflect.String:
					tags := strings.ReplaceAll(value.(string), ";", ",")
					tagsArray := strings.Split(tags, ",")
					for i, tag := range tagsArray {
						tag = strings.Trim(tag, " ")
						tagsArray[i] = tag
					}
					article.Tags = tagsArray
				}
			}
		}
	}

	// Set Created and Updated to file dates if not provided in frontmatter.
	fileInfo, err := os.Stat(path)
	if err != nil {
		return Article{}, fmt.Errorf("failed to get file info: %w", err)
	}
	if article.Created.IsZero() {
		article.Created = DateTimeFromString(path) // Try to extract date from filename
		if article.Created.IsZero() {
			article.Created = fileInfo.ModTime() // Use file modification time
		}
	}
	if article.Updated.IsZero() {
		article.Updated = fileInfo.ModTime() // Use file modification time
	}

	// Default title to filename if not provided.
	if article.Title == "" {
		article.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	// Extract resources from HTML (commented out).
	// article.Files = extractResources(content) // Pass content here, not article.Content

	// Determine if the article is a page (commented out).
	// article.IsPage = contains(article.Tags, "PAGE")

	// Set the article path (already done).
	// article.OriginalPath = filepath.Dir(path)

	// Set the HTML content (already done).
	// article.HtmlContent = content

	return article, nil
}

// FormatMarkdown applies an HTML template to the Markdown content of an article.
func FormatMarkdown(article *Article, settings Settings) {
	// Define template functions.
	tmpl, err := template.New("markdown_template").Funcs(
		template.FuncMap{
			"stringsJoin":    strings.Join,
			"slicesContains": slices.Contains[[]string]}).Parse(htmlArticleTemplate)
	if err != nil {
		panic(err) // Potential Problem: Panicking here will crash the application. Consider returning an error.
	}

	// Execute the template with article data and settings.
	var tp bytes.Buffer
	err = tmpl.Execute(&tp, struct {
		Art      Article
		Ctt      template.HTML
		Settings Settings
	}{*article, template.HTML(article.HtmlContent), settings})
	if err != nil {
		panic(err) // Potential Problem: Panicking here will crash the application. Consider returning an error.
	}
	article.HtmlContent = tp.String()
}

// HTMLFile parses an HTML file, extracts metadata from tags, and populates an Article struct.
func HTMLFile(path string) (Article, error) {
	// Read the HTML file content.
	data, err := os.ReadFile(path)
	if err != nil {
		return Article{}, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	htmlContent := string(data)
	textContent := html2text.HTML2Text(htmlContent)

	// Create an Article struct with basic information.
	article := Article{
		OriginalPath: path,
		HtmlContent:  htmlContent,
		TextContent:  textContent,
	}
	// Parse the HTML content.
	htmlTree, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return Article{}, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Get info from <title> tag.
	titleNode := findFirstElement(htmlTree, "title")
	if titleNode != nil && titleNode.FirstChild != nil {
		article.Title = titleNode.FirstChild.Data
	}

	// Default title to filename if not found in <title> tag.
	if article.Title == "" {
		article.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	// Get info from meta tags.
	for _, metaTag := range findAllElements(htmlTree, "meta") {
		key := ""
		val := ""
		for _, attr := range metaTag.Attr {
			attrKey := strings.ToLower(attr.Key)
			attrKey = strings.Trim(attrKey, " ")
			switch attrKey {
			case "name":
				key = attr.Val
			case "content":
				val = attr.Val
			}
		}
		// Populate Article fields based on meta tag content.
		if key != "" && val != "" {
			switch key {
			case "description":
				article.Description = val
			case "keywords":
				tags := strings.ReplaceAll(val, ";", ",")
				tagsArray := strings.Split(tags, ",")
				for i, tag := range tagsArray {
					tag = strings.Trim(tag, " ")
					tagsArray[i] = tag
				}
				article.Tags = tagsArray
			case "created":
				article.Created = DateTimeFromString(val)
			case "updated":
				article.Updated = DateTimeFromString(val)
			}
		}
	}

	// Set Created and Updated to file dates if not provided in meta tags.
	fileInfo, err := os.Stat(path)
	if err != nil {
		return Article{}, fmt.Errorf("failed to get file info: %w", err)
	}
	if article.Created.IsZero() {
		article.Created = DateTimeFromString(path) // Try to extract date from filename
		if article.Created.IsZero() {
			article.Created = fileInfo.ModTime() // Use file modification time
		}
	}
	if article.Updated.IsZero() {
		article.Updated = fileInfo.ModTime() // Use file modification time
	}

	return article, nil
}

// findFirstElement recursively searches for the first HTML element with the given tag name.
func findFirstElement(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirstElement(c, tag); found != nil {
			return found
		}
	}
	return nil
}

// findAllElements recursively searches for all HTML elements with the given tag name.
func findAllElements(n *html.Node, tag string) []*html.Node {
	var elements []*html.Node
	if n.Type == html.ElementNode && n.Data == tag {
		elements = append(elements, n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		elements = append(elements, findAllElements(c, tag)...)
	}
	return elements
}

// extractResources parses HTML content and extracts the values of "src" and "href" attributes
// from "img", "script", and "link" tags, returning a list of resource paths.
func extractResources(htmlContent string) []string {
	var resources []string
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		panic(err) // Potential Problem: Panicking here will crash the application. Consider returning an error.
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
