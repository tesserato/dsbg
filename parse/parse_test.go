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
tags: {tags}
created: {created}
updated: {updated}
---

# header 1

paragraph beneath header 1

## header 2

paragraph beneath header 1
`

// type markdown_tests_out struct {
// 	path        string
// 	title       string
// 	description string
// 	tags        []string
// 	created     time.Time
// 	updated     time.Time
// }

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
		path:        "2023-05-11-test-entry.md",
		title:       "Title",
		description: "Description",
		tags:        "[tag1, tag2, compound tag1, compound tag2]",
		created:     "2023-05-11",
		updated:     "2024-06-13",
		tagsOut:     []string{"tag1", "tag2", "compound tag1", "compound tag2"},
		createdOut:  time.Date(2023, 5, 11, 0, 0, 0, 0, time.UTC),
		updatedOut:  time.Date(2024, 6, 13, 0, 0, 0, 0, time.UTC),
	},
}

func genMarkdownText(template string, aStruct any) string {
	fields := reflect.VisibleFields(reflect.TypeOf(aStruct))
	vals := reflect.Indirect(reflect.ValueOf(aStruct))
	for _, field := range fields {
		val := vals.FieldByName(field.Name)
		// val := fields.FieldByNamv.FieldByIndex(e.Index)
		fmt.Printf("Key: %s\tType: %s \t Value: %s\n", field.Name, field.Type, val)

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

		path := dir + "/" + e.path
		os.WriteFile(path, []byte(mdText), 0644)
		// fmt.Println(path)

		md, err := MarkdownFile(path)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if md.Title != e.title {
			t.Errorf("Wrong title: %s != %s", e.title, md.Title)
		}

		if md.Description != e.description {
			t.Errorf("Wrong description: %s != %s",  e.description, md.Description)
		}

		if !compareTags(md.Tags, e.tagsOut) {
			t.Errorf("Wrong tags: %v != %v", e.tagsOut, md.Tags)
		}

		if !md.Created.Equal(e.createdOut) {
			t.Errorf("Wrong created date: %s != %s", e.createdOut, md.Created)
		}

		if !md.Updated.Equal(e.updatedOut) {
			t.Errorf("Wrong updated date: %s != %s", e.updatedOut, md.Updated)
		}
	}
}
