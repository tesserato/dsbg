package parse

import (
	"html/template"
	"time"
)

type Style int64

const (
	Default Style = iota
	Dark
	Colorful
)

type Settings struct {
	Title                     string
	Description               string
	InputDirectory            string
	OutputDirectory           string
	DateFormat                string
	IndexName                 string
	Style                     Style
	PathToCustomCss           string
	PathToCustomJs            string
	PathToCustomFavicon       string
	AdditionalElementsTop     template.HTML
	AdditionalElemensBottom   template.HTML
	DoNotExtractTagsFromPaths bool
	DoNotRemoveDateFromPaths  bool
	DoNotRemoveDateFromTitles bool
	OpenInNewTab              bool
	BaseUrl                   string
}

type Article struct {
	Title        string
	Description  string
	Created      time.Time
	Updated      time.Time
	Tags         []string
	TextContent  string
	HtmlContent  string
	OriginalPath string
	LinkToSelf   string
	LinkToSave   string
}

// type Links struct {
// 	ToSelf string
// 	// ToCss  string
// 	// ToJs   string
// 	ToSave string
// }
