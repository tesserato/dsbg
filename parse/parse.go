package parse

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"slices"

	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
	"golang.org/x/net/html"
)

type Settings struct {
	Title           string
	InputDirectory  string
	OutputDirectory string
	DateFormat      string
	IndexName       string
}

func NewSettings() Settings {
	settings := Settings{}
	settings.Title = "Blog"
	settings.InputDirectory = "content"
	settings.OutputDirectory = "public"
	settings.DateFormat = "2006-01-02"
	settings.IndexName = "index.html"
	return settings
}

type Article struct {
	Title        string
	Description  string
	Created      time.Time
	Updated      time.Time
	Tags         []string
	HtmlContent  string
	OriginalPath string
	LinkToSelf   string
	// LinkToCss    string
	// LinkToJs     string
	// InnerHTML string
}

func DateTimeFromString(date string) time.Time {
	m := make(map[string]int)
	for _, pattern := range []string{
		`(?P<year>\d{4})\D+(?P<month>\d{1,2})\D+(?P<day>\d{1,2})`,
		`(?P<day>\d{1,2})\D+(?P<month>\d{1,2})\D+(?P<year>\d{4})`,
		`(?P<hour>\d{2}):(?P<min>\d{2}):(?P<sec>\d{2})`,
	} {
		r := regexp.MustCompile(pattern)

		matches := r.FindStringSubmatch(date)
		if len(matches) > 0 {
			for i, name := range r.SubexpNames()[1:] {
				integer, err := strconv.Atoi(matches[i+1])
				if err != nil {
					panic(err) // TODO handle error
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

var htmlArticleTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{.Art.Title}}</title>
	<link rel="stylesheet" href="{{.PathToCss}}">
	<script src="{{.PathToJs}}"></script>
</head>
<body>
	<article>
		<h1>{{.Art.Title}}</h1>

		{{if slicesContains .Art.Tags "PAGE"}}
			<p>Created: {{.Art.Created}}</p>
			<p>Updated: {{.Art.Updated}}</p>
			<ul>
				{{.Art.Tags}}
			</ul>
		{{end}}
		{{.Ctt}}
	</article>
</body>
</html>
`

func MarkdownFile(path string, pathToCss string, pathToJs string) (Article, error) {
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
	content := buf.String()

	// Retrieve frontmatter from the context
	var article = Article{OriginalPath: path}
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

	tmpl, err := template.New("markdown_template").Funcs(template.FuncMap{"stringsJoin": strings.Join, "slicesContains": slices.Contains[[]string]}).Parse(htmlArticleTemplate)
	if err != nil {
		panic(err)
	}

	var tp bytes.Buffer
	err = tmpl.Execute(&tp, struct {
		Art Article
		Ctt template.HTML
		PathToCss string
		PathToJs  string
	}{article, template.HTML(content), pathToCss, pathToJs})
	if err != nil {
		panic(err)
	}
	htmlContent := tp.String()
	article.HtmlContent = htmlContent

	return article, nil
}

func HTMLFile(path string) (Article, error) {
	// Read the HTML file content
	data, err := os.ReadFile(path)
	if err != nil {
		return Article{}, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	content := string(data)

	// Create an article and populate common fields
	article := Article{
		OriginalPath: path,
		HtmlContent:  content,
		// Files:        extractResources(content),
	}
	htmlTree, err := html.Parse(strings.NewReader(content))
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

// Helper function to extract text content from an HTML node
// func getTextContent(n *html.Node) string {
// 	var text string
// 	if n.Type == html.TextNode {
// 		text += n.Data
// 	}
// 	for c := n.FirstChild; c != nil; c = c.NextSibling {
// 		text += getTextContent(c)
// 	}
// 	return strings.TrimSpace(text)
// }

// func contains(s []string, e string) bool {
// 	for _, a := range s {
// 		if a == e {
// 			return true
// 		}
// 	}
// 	return false
// }

// func SaveHtml(article Article, settings Settings) error {
// 	// Create the page folder if it doesn't exist
// 	pageDir := filepath.Join(outputDir, a.OriginalPath)
// 	err := os.MkdirAll(pageDir, 0755)
// 	if err != nil {
// 		return err
// 	}

// 	// resources := extractResources(a.HtmlContent)

// 	// Sanitize the HTML content
// 	// p := bluemonday.UGCPolicy()
// 	// html = p.Sanitize(html)

// 	// Write the HTML content to the file
// 	filename := filepath.Join(pageDir, defaultIndexName)
// 	return os.WriteFile(filename, []byte(a.HtmlContent), 0644)
// }

// func GenerateTagsHTML(tags []string) string {
// 	var html string
// 	for _, tag := range tags {
// 		html += fmt.Sprintf("<li>%s</li>\n", tag)
// 	}
// 	return html
// }

func ExtractResources(htmlContent string) []string {
	var resources []string
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		fmt.Errorf("error parsing HTML: %w", err)
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
