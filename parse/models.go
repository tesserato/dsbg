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
	Description               template.HTML
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
	Title          string
	Description    string
	CoverImagePath string
	Created        time.Time
	Updated        time.Time
	Tags           []string
	TextContent    string
	HtmlContent    string
	OriginalPath   string
	LinkToSelf     string
	LinkToSave     string
}

type Theme struct {
	HeaderFont   string
	BodyFont     string
	Background   string
	Text         string
	Card         string
	Link         string
	Shadow       string
	Button       string
}