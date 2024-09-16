package parse

import (
	"fmt"
	"os"
	"testing"
	"time"
)

type markdown_tests_out struct {
	path        string
	title       string
	description string
	tags        []string
	created     time.Time
	updated     time.Time
}

var markdown_tests = []struct {
	in  string
	out markdown_tests_out
}{
	{
		`---
title: Title
description: Description
tags: [tag1, tag2, compound tag1, compound tag2]
created: 2023-05-11
updated: 2024-06-13
---

# header 1

paragraph beneath header 1

## header 2

paragraph beneath header 1`,
		markdown_tests_out{
			"2023-05-11-test-entry.md",
			"Title",
			"Description",
			[]string{"tag1", "tag2", "compound tag1", "compound tag2"},
			time.Date(2023, 5, 11, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 6, 13, 0, 0, 0, 0, time.UTC),
		},
	},
	{
		`---
title: Another Title
description: Another Description
tags: [tag3, tag4, another compound tag]
created: 2022-12-01
updated: 2023-01-20
---

# header 1

A different paragraph beneath header 1

## header 2

Another paragraph beneath header 1`,
		markdown_tests_out{
			"2022-12-01-another-test-entry.md",
			"Another Title",
			"Another Description",
			[]string{"tag3", "tag4", "another compound tag"},
			time.Date(2022, 12, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2023, 1, 20, 0, 0, 0, 0, time.UTC),
		},
	},
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

	for _, e := range markdown_tests {
		path := dir + "/" + e.out.path
		os.WriteFile(path, []byte(e.in), 0644)
		fmt.Println(path)

		md, err := MarkdownFile(path)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if md.Title != e.out.title {
			t.Errorf("Wrong title: %s != %s", md.Title, e.out.title)
		}

		if md.Description != e.out.description {
			t.Errorf("Wrong description: %s != %s", md.Description, e.out.description)
		}

		if !compareTags(md.Tags, e.out.tags) {
			t.Errorf("Wrong tags: %v != %v", md.Tags, e.out.tags)
		}

		if !md.Created.Equal(e.out.created) {
			t.Errorf("Wrong created date: %s != %s", md.Created, e.out.created)
		}

		if !md.Updated.Equal(e.out.updated) {
			t.Errorf("Wrong updated date: %s != %s", md.Updated, e.out.updated)
		}
	}
}
