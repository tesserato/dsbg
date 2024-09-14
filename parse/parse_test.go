package parse

import (
	"os"
	"testing"
)

var md1 = `---
title: Title
description: Description
tags: tag1, tag2, compound tag1, compound tag2
---

# header 1

paragraph beneath header 1

## header 2

paragraph beneath header 1
`

func TestMarkdownFile(t *testing.T) {
	dir := t.TempDir()

	filename := dir + "/2016 05 29 Test entry.md"
	os.WriteFile(filename, []byte(md1), 0644)
}
