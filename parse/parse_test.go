package parse

import (
	"fmt"
	"os"
	"testing"
)

type markdown_tests_out struct {
	path        string
	title       string
	description string
	tags        []string
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
created: 2024 5 11
updated: 2024 06 13
---

# header 1

paragraph beneath header 1

## header 2

paragraph beneath header 1`,
		markdown_tests_out{"2016 05 29 Test entry.md", "Title", "Description", []string{"tag1", "tag2", "compound tag1", "compound tag2"}},
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
		fmt.Println(e.in)

		md, err := MarkdownFile(path)
		if err != nil {
			t.Fatalf("Error")
		}

		if md.Description != e.out.description {
			t.Errorf("Wrong description")
		}

		if !compareTags(md.Tags, e.out.tags) {
			t.Errorf("Wrong tags: %v != %v", md.Tags, e.out.tags)
		}
	}

}
