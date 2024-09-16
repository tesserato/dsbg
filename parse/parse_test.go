package parse

import (
	"fmt"
	"os"
	"testing"
)

var markdown_tests = []struct {
	in  string
	out string
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
		" "},
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

	path := dir + "/2016 05 29 Test entry.md"

	fmt.Println(md1)
	os.WriteFile(path, []byte(md1), 0644)

	md, err := MarkdownFile(path)
	if err != nil {
		t.Fatalf("Error")
	}

	if md.Description != "Description" {
		t.Errorf("Wrong description")
	}

	tags := []string{"tag1", "tag2", "compound tag1", "compound tag2"}
	if !compareTags(tags, md.Tags) {
		t.Errorf("Wrong tags: %v != %v", tags, md.Tags)
	}
}
