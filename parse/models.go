package parse

import (
	"html/template"
	"time"
)

type Settings struct {
	Title                   string
	InputDirectory          string
	OutputDirectory         string
	DateFormat              string
	IndexName               string
	AdditionalElementsTop   template.HTML
	AdditionalElemensBottom template.HTML
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

type Links struct {
	ToSelf string
	ToCss  string
	ToJs   string
	ToSave string
}
