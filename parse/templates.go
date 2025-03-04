package parse

var FrontMatterTemplate = `---
title: {{.Title}}
description: {{.Description}}
created: {{.Created}}
updated: {{.Updated}}
coverImagePath: {{.CoverImagePath}}
tags: {{.Tags}}
---

# Markdown Content Examples

Here's a mix of common Markdown elements to show you how they look:

## Lists

- An unordered list
    - Nested list item
- Another list item

1.  An ordered list
2.  Second item
    1.  Nested ordered list

## Code

You can include code blocks with syntax highlighting:

## Go example:
` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("Hello, DSBG!")
}
` + "```" + `

## Rust example:
` + "```rust" + `
fn main() {
    println!("Hello, DSBG!");
}
` + "```" + `

## Python example:
` + "```python" + `
def main():
    print("Hello, DSBG!")

if __name__ == "__main__":
    main()
` + "```" + `

## JavaScript example:
` + "```javascript" + `
function main() {
  console.log("Hello, DSBG!");
}

main();
` + "```" + `

## Images

Images are automatically copied to the output directory:

![A Butterfly](01_butterfly.webp)

## Text Formatting

Here's some **bold** and *italic* text, and some ` + "`inline code`" + `.

## Links and Quotes

Check out this [example link](https://example.com).

> This is a block quote. Useful for citing resources.
`

const rssTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"
    xmlns:atom="http://www.w3.org/2005/Atom"
    xmlns:content="http://purl.org/rss/1.0/modules/content/"
    xmlns:media="http://search.yahoo.com/mrss/">
	<channel>
		<title>{{ .Settings.Title }}</title>
		<link>{{ .Settings.BaseUrl }}</link>
        <atom:link href="{{ .Settings.BaseUrl }}/rss.xml" rel="self" type="application/rss+xml" />
		<description>{{ htmlEscape .Settings.DescriptionMarkdown }}</description>
		<generator>Dead Simple Blog Generator</generator>
		<lastBuildDate>{{ .BuildDate }}</lastBuildDate>
		{{- range .Articles }}
		<item>
			<title>{{ .Title | htmlEscape }}</title>
			<link>{{ buildArticleURL . $.Settings  }}</link>
			<guid>{{ buildArticleURL . $.Settings  }}</guid>
			<pubDate>{{ .Created | formatPubDate }}</pubDate>
			<description>{{ .Description | htmlEscape }}</description>
			{{- if .CoverImagePath }}
			<media:content
				url="{{ $.Settings.BaseUrl}}/{{ .CoverImagePath }}"
				medium="image"
				type="image/jpeg"
			/>
			{{- end }}
            {{- range .Tags }}
            <category>{{ . | htmlEscape }}</category>
            {{- end }}
		</item>
		{{- end }}
	</channel>
</rss>
`
