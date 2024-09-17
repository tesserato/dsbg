package parse

import (
	"fmt"
	// "strconv"
	"strings"

	"os"
	"reflect"
	"testing"
	"time"
)

var mdTemplate string = `---
title: {title}
description: {description}
tags: [{tags}]
created: {created}
updated: {updated}
---

# header 1

paragraph beneath header 1

## header 2

paragraph beneath header 1
`
var htmlTemplate string = `<!DOCTYPE html>
<html lang="en">

<head>
	<title>{title}</title>
	<meta name="description" content="{description}">
	<meta name="keywords" content= "{tags}" >
	<meta name="created" content="{created}">
	<meta name="updated" content="{updated}">

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
		"2014 06 06 How to set up a free blog.md",
	} {
		tm := DateTimeFromString(d)
		if tm != defaultTime {
			t.Errorf("Wrong date: %s != %s", defaultTime, tm)
		}
	}
}

var dataTestMarkdownFile = []struct {
	path        string
	title       string
	description string
	tags        string
	created     string
	updated     string
	tagsOut     []string
	createdOut  time.Time
	updatedOut  time.Time
}{
	{
		path:        "2023-05-11-test-entry",
		title:       "Title",
		description: "Description",
		tags:        "tag1, tag2, compound tag1, compound tag2",
		created:     "2020-05-11",
		updated:     "2021 06-13",
		tagsOut:     []string{"tag1", "tag2", "compound tag1", "compound tag2"},
		createdOut:  time.Date(2020, 5, 11, 0, 0, 0, 0, time.UTC),
		updatedOut:  time.Date(2021, 6, 13, 0, 0, 0, 0, time.UTC),
	},
	{
		path:        "2023-05-11-test-entry",
		title:       "Title",
		description: "Description",
		tags:        "just this one compound tag",
		// created:     "2020-05-11",
		updated:    "13 06-2021",
		tagsOut:    []string{"just this one compound tag"},
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
	for _, e := range dataTestMarkdownFile {

		mdText := genMarkdownText(mdTemplate, e)
		fmt.Println(mdText)

		path := dir + "/" + e.path + ".md"
		os.WriteFile(path, []byte(mdText), 0644)

		md, err := MarkdownFile(path)
		fmt.Println(md.HtmlContent)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if md.Title != e.title {
			t.Errorf("Wrong title: %s != %s", e.title, md.Title)
		}

		if md.Description != e.description {
			t.Errorf("Wrong description: %s != %s", e.description, md.Description)
		}

		if !compareTags(md.Tags, e.tagsOut) {
			t.Errorf("Wrong tags: %v != %v", e.tagsOut, md.Tags)
		}

		if !md.Created.Equal(e.createdOut) {
			t.Errorf("Wrong created date: (%s) %s != %s", e.created, e.createdOut, md.Created)
		}

		if !md.Updated.Equal(e.updatedOut) {
			t.Errorf("Wrong updated date: (%s) %s != %s", e.updated, e.updatedOut, md.Updated)
		}
	}
}

func TestHTMLFile(t *testing.T) {
	dir := t.TempDir()
	for _, e := range dataTestMarkdownFile {

		htmlText := genMarkdownText(htmlTemplate, e)
		fmt.Println(htmlText)
		
		path := dir + "/" + e.path + ".html"
		os.WriteFile(path, []byte(htmlText), 0644)

		html, err := HTMLFile(path)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if html.Title != e.title {
			t.Errorf("Wrong title: %s != %s", e.title, html.Title)
		}

		if html.Description != e.description {
			t.Errorf("Wrong description: %s != %s", e.description, html.Description)
		}

		if !compareTags(html.Tags, e.tagsOut) {
			t.Errorf("Wrong tags: %v != %v", e.tagsOut, html.Tags)
		}

		if !html.Created.Equal(e.createdOut) {
			t.Errorf("Wrong created date: (%s) %s != %s", e.created, e.createdOut, html.Created)
		}

		if !html.Updated.Equal(e.updatedOut) {	
			t.Errorf("Wrong updated date: (%s) %s != %s", e.updated, e.updatedOut, html.Updated)
		}
	}
}