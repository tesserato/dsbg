package parse

import (
	"bytes"
	"fmt"
	"html/template"

	// "strconv"
	"strings"

	"os"
	"reflect"
	"testing"
	"time"
)

var mdTemplate string = `---
title: {{.Title}}
description: {{.Description}}
tags: {{stringsJoin .Tags ", "}}
created: {{.Created}}
updated: {{.Updated}}
---

# header 1

paragraph beneath header 1

## header 2

paragraph beneath header 1
`
var htmlTemplate string = `<!DOCTYPE html>
<html lang="en">

<head>
	<title>{{.Title}}</title>
	<meta name="description" content="{{.Description}}">
	<meta name="keywords" content= "{{stringsJoin .Tags ", "}}" >
	<meta name="created" content="{{.Created}}">
	<meta name="updated" content="{{.Updated}}">

	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<link rel="stylesheet" href="style.css" />
</head>

<body>
	<h1 id="header-1">header 1</h1>
	<p>paragraph beneath header 1</p>
	<h2 id="header-2">header 2</h2>
	<p>paragraph beneath header 1</p>
</body>
`

func TestDateTimeFromString(t *testing.T) {
	defaultTime := time.Date(2014, 9, 6, 0, 0, 0, 0, time.UTC)
	for _, d := range []string{
		"2014 09 06",
		"2014 9 06",
		"2014-9-06",
		"2014 09 6",
		"6 9 2014",
		"6   9   2014",
		"6/9 2014",
		"6.9.2014",
		"6_9-2014",
		"2014 09 06 How to set up a free blog.md",
	} {
		tm := DateTimeFromString(d)
		if tm != defaultTime {
			t.Errorf("Wrong date: %s != %s", defaultTime, tm)
		}
	}
}

var dataTestMarkdownFile = []struct {
	Path        string
	Title       string
	Description string
	Tags        []string
	Created     string
	Updated     string
	createdOut  time.Time
	updatedOut  time.Time
}{
	{
		Path:        "2023-05-11-test-entry",
		Title:       "Title",
		Description: "Description",
		Tags:        []string{"tag1", "tag2", "compound tag1", "compound tag2"},
		Created:     "2020-05-11",
		Updated:     "2021 06-13",
		createdOut:  time.Date(2020, 5, 11, 0, 0, 0, 0, time.UTC),
		updatedOut:  time.Date(2021, 6, 13, 0, 0, 0, 0, time.UTC),
	},
	{
		Path:        "2023-05-11-test-entry",
		Title:       "Title",
		Description: "Description",
		Tags:        []string{"just this one compound tag"},
		// created:     "2020-05-11",
		Updated:    "13 06-2021",
		createdOut: time.Date(2023, 5, 11, 0, 0, 0, 0, time.UTC),
		updatedOut: time.Date(2021, 6, 13, 0, 0, 0, 0, time.UTC),
	},
}

func genMarkdownText(template string, aStruct any) string {
	fields := reflect.VisibleFields(reflect.TypeOf(aStruct))
	values := reflect.Indirect(reflect.ValueOf(aStruct))
	for _, field := range fields {
		val := values.FieldByName(field.Name)
		// val := fields.FieldByNamv.FieldByIndex(e.Index)
		// fmt.Printf("Key: %s\tType: %s \t Value: %s\n", field.Name, field.Type, val)

		// if field.Type.Kind() == reflect.String {
		// val := val.String()

		// }
		template = strings.ReplaceAll(template, "{"+field.Name+"}", val.String())
	}
	return template
}

func compareTags(tags1 []string, tags2 []string) bool {
	if len(tags1) != len(tags2) {
		return false
	}
	for i, v := range tags1 {
		if v != tags2[i] {
			return false
		}
	}
	return true
}

func TestMarkdownFile(t *testing.T) {
	dir := t.TempDir()
	tmpl, err := template.New("markdown_template").Funcs(template.FuncMap{"stringsJoin": strings.Join}).Parse(mdTemplate)
	if err != nil {
		panic(err)
	}

	for _, e := range dataTestMarkdownFile {

		var tp bytes.Buffer
		err = tmpl.Execute(&tp, e)
		if err != nil {
			panic(err)
		}
		mdText := tp.String()
		fmt.Println(mdText)

		path := dir + "/" + e.Path + ".md"
		os.WriteFile(path, []byte(mdText), 0644)

		md, err := MarkdownFile(path)
		fmt.Println(md.HtmlContent)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if md.Title != e.Title {
			t.Errorf("Wrong title: %s != %s", e.Title, md.Title)
		}

		if md.Description != e.Description {
			t.Errorf("Wrong description: %s != %s", e.Description, md.Description)
		}

		if !compareTags(md.Tags, e.Tags) {
			t.Errorf("Wrong tags: %v != %v", e.Tags, md.Tags)
		}

		if !md.Created.Equal(e.createdOut) {
			t.Errorf("Wrong created date: (%s) %s != %s", e.Created, e.createdOut, md.Created)
		}

		if !md.Updated.Equal(e.updatedOut) {
			t.Errorf("Wrong updated date: (%s) %s != %s", e.Updated, e.updatedOut, md.Updated)
		}
	}
}

func TestHTMLFile(t *testing.T) {
	dir := t.TempDir()
	tmpl, err := template.New("markdown_template").Funcs(template.FuncMap{"stringsJoin": strings.Join}).Parse(htmlTemplate)
	if err != nil {
		panic(err)
	}
	for _, e := range dataTestMarkdownFile {

		var tp bytes.Buffer
		err = tmpl.Execute(&tp, e)
		if err != nil {
			panic(err)
		}
		htmlText := tp.String()
		fmt.Println(htmlText)

		path := dir + "/" + e.Path + ".html"
		os.WriteFile(path, []byte(htmlText), 0644)

		html, err := HTMLFile(path)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if html.Title != e.Title {
			t.Errorf("Wrong title: %s != %s", e.Title, html.Title)
		}

		if html.Description != e.Description {
			t.Errorf("Wrong description: %s != %s", e.Description, html.Description)
		}

		if !compareTags(html.Tags, e.Tags) {
			t.Errorf("Wrong tags: %v != %v", e.Tags, html.Tags)
		}

		if !html.Created.Equal(e.createdOut) {
			t.Errorf("Wrong created date: (%s) %s != %s", e.Created, e.createdOut, html.Created)
		}

		if !html.Updated.Equal(e.updatedOut) {
			t.Errorf("Wrong updated date: (%s) %s != %s", e.Updated, e.updatedOut, html.Updated)
		}
	}
}
