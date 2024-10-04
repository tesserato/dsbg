package parse

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"

	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/k3a/html2text"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
	"golang.org/x/net/html"
)

var regexPatterns = []string{
	`(?P<year>\d{4})\D+(?P<month>\d{1,2})\D+(?P<day>\d{1,2})`,
	`(?P<day>\d{1,2})\D+(?P<month>\d{1,2})\D+(?P<year>\d{4})`,
	`(?P<hour>\d{2}):(?P<min>\d{2}):(?P<sec>\d{2})`,
}

func RemoveDateFromPath(stringWithDate string) string {
	fmt.Printf("?? RemoveDateFromPath: %s\n", stringWithDate)
	for _, pattern := range regexPatterns {
		r := regexp.MustCompile(pattern)
		stringWithDate = r.ReplaceAllString(stringWithDate, "")
	}
	fmt.Printf("?? RemoveDateFromPath: %s\n", stringWithDate)
	return stringWithDate
}

func DateTimeFromString(date string) time.Time {
	m := make(map[string]int)
	for _, pattern := range regexPatterns {
		r := regexp.MustCompile(pattern)
		matches := r.FindStringSubmatch(date)
		if len(matches) > 0 {
			for i, name := range r.SubexpNames()[1:] {
				integer, err := strconv.Atoi(matches[i+1])
				if err != nil {
					panic(err)
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
func GetPaths(root string, extensions []string) ([]string, error) {
	for i, ext := range extensions {
		ext = strings.Trim(ext, " ")
		extensions[i] = strings.ToLower(ext)
	}
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			for _, ext := range extensions {
				path = strings.ToLower(path)
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

func cleanString(url string) string {
	fmt.Printf("!! cleanString: %s\n", url)
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
	fmt.Printf("!! cleanString: %s\n", url)
	return url
}

func CopyHtmlResources(settings Settings, originalArticlePath string, htmlContent string) Links {
	relativeInputPath, err := filepath.Rel(settings.InputDirectory, originalArticlePath)
	if err != nil {
		panic(err)
	}
	// outputPath
	outputPath := filepath.Join(settings.OutputDirectory, relativeInputPath)
	outputPath = strings.TrimSuffix(outputPath, filepath.Ext(outputPath))
	outputPath = filepath.Join(outputPath, settings.IndexName)
	if !settings.DoNotRemoveDateFromPaths {
		outputPath = RemoveDateFromPath(outputPath)
	}
	outputPath = cleanString(outputPath)

	outputDirectory := filepath.Dir(outputPath)

	err = os.MkdirAll(outputDirectory, os.ModePerm)
	if err != nil {
		panic(err)
	}

	originalDirectory := filepath.Dir(originalArticlePath)
	for _, resourceOrigRelPath := range extractResources(htmlContent) {
		resourceOrigRelPathLower := strings.ToLower(resourceOrigRelPath)
		if strings.Contains(resourceOrigRelPathLower, "http") {
			continue
		}
		resourceOrigPath := filepath.Join(originalDirectory, resourceOrigRelPath)
		resourceDestPath := filepath.Join(outputDirectory, resourceOrigRelPath)
		fmt.Printf("  resourceOrigPath: %s\n  resourceDestPath: %s\n\n", resourceOrigPath, resourceDestPath)

		input, err := os.ReadFile(resourceOrigPath)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(resourceDestPath, input, 0644)
		if err != nil {
			panic(err)
		}
	}

	LinkToSelf, err := filepath.Rel(settings.OutputDirectory, outputPath)
	if err != nil {
		panic(err)
	}
	LinkToSelf = strings.ReplaceAll(LinkToSelf, "\\", "/")
	fmt.Printf(
		"InputDirectory: %s\noriginalArticlePath: %s\nrelativeInputPath: %s\noutputDirectory: %s\noutputPath: %s\nLinkToSelf: %s\n\n",
		settings.InputDirectory, originalArticlePath, relativeInputPath, outputDirectory, outputPath, LinkToSelf)

	return Links{ToSelf: LinkToSelf, ToSave: outputPath}
}

func GenerateHtmlIndex(articles []Article, settings Settings) error {
	// Generate the article list HTML
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
		PageList    []Article
		ArticleList []Article
		Settings    Settings
	}{allTags, pageList, articleList, settings})
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	// Write the HTML content to the file
	filePath := filepath.Join(settings.OutputDirectory, settings.IndexName)
	return os.WriteFile(filePath, tp.Bytes(), 0644)
}

func MarkdownFile(path string) (Article, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Article{}, err
	}

	markdown := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAttribute(),
			parser.WithAutoHeadingID(),
		),
		goldmark.WithExtensions(
			&frontmatter.Extender{},
		),
	)

	// Create a context to store frontmatter
	context := parser.NewContext()

	// Parse the Markdown file, storing frontmatter in the context
	var buf strings.Builder
	if err := markdown.Convert(data, &buf, parser.WithContext(context)); err != nil {
		panic(err)
	}
	// content := buf.String()

	// Retrieve frontmatter from the context
	var article = Article{OriginalPath: path, TextContent: string(data), HtmlContent: buf.String()}
	fm := frontmatter.Get(context)
	if fm != nil {
		var d map[string]any
		if err := fm.Decode(&d); err != nil {
			fmt.Printf("Article path: %s", path)
			panic(err)
		}
		for name, value := range d {
			// fmt.Printf("Key: %s\tValue: %v\tType: %T\n", name, value, value)
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

	// 2. Set Created and Updated to file dates if not provided in frontmatter
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

	// 3. Default title to filename if not provided
	if article.Title == "" {
		article.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	// Extract resources from HTML
	// article.Files = extractResources(content) // Pass content here, not article.Content

	// Determine if the article is a page
	// article.IsPage = contains(article.Tags, "PAGE")

	// Set the article path
	// article.OriginalPath = filepath.Dir(path)

	// article.HtmlContent = content

	return article, nil
}

func FormatMarkdown(article Article, links Links, settings Settings) Article {
	tmpl, err := template.New("markdown_template").Funcs(
		template.FuncMap{
			"stringsJoin":    strings.Join,
			"slicesContains": slices.Contains[[]string]}).Parse(htmlArticleTemplate)
	if err != nil {
		panic(err)
	}

	var tp bytes.Buffer
	err = tmpl.Execute(&tp, struct {
		Art      Article
		Ctt      template.HTML
		Lks      Links
		Settings Settings
	}{article, template.HTML(article.HtmlContent), links, settings})
	if err != nil {
		panic(err)
	}
	htmlContent := tp.String()
	article.HtmlContent = htmlContent
	article.LinkToSelf = links.ToSelf
	return article
}

func HTMLFile(path string) (Article, error) {
	// Read the HTML file content
	data, err := os.ReadFile(path)
	if err != nil {
		return Article{}, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	htmlContent := string(data)
	textContent := html2text.HTML2Text(htmlContent)

	// Create an article and populate common fields
	article := Article{
		OriginalPath: path,
		HtmlContent:  htmlContent,
		TextContent:  textContent,
	}
	htmlTree, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return Article{}, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Get info from <title> tag
	article.Title = findFirstElement(htmlTree, "title").FirstChild.Data

	if article.Title == "" {
		article.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	// Get info from meta tags
	for _, metaTag := range findAllElements(htmlTree, "meta") {
		key := ""
		val := ""
		for _, attr := range metaTag.Attr {
			// fmt.Printf("Key: %s\tValue: %s\n", attr.Key, attr.Val)
			attrKey := strings.ToLower(attr.Key)
			attrKey = strings.Trim(attrKey, " ")
			switch attrKey {
			case "name":
				key = attr.Val
			case "content":
				val = attr.Val
			}
		}
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

	// Set Created and Updated to file dates if not provided in frontmatter
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

// Helper function to find the first occurrence of an element by tag name
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

func extractResources(htmlContent string) []string {
	var resources []string
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		panic(err)
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
