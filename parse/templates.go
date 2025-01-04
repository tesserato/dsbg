package parse

var FrontMatterTemplate = `---
title: {{.Title}}
description: {{.Description}}
created: {{.CurrentDate}}
updated: {{.CurrentDate}}
coverImagePath: 
tags:
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

var htmlArticleTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	{{.Settings.AdditionalElementsTop}}
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="generator" content="ZBSCM">
	<meta name="keywords" content="{{stringsJoin .Art.Tags ", "}}">
    <link rel="stylesheet" href="/style.css">
    <link rel="icon" type="image/x-icon" href="/favicon.ico">

    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/default.min.css">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
    <script>hljs.highlightAll();</script>

    <title>{{.Art.Title}}</title>
</head>

<body>
    <header>
		<a href="/"> &#9665; {{.Settings.Title}} </a>
        <h1>{{.Art.Title}}</h1>
        <h2>{{.Art.Description}}</h2>
    </header>
    <div class="detail">
        {{.Ctt}}
    </div>
    <div class="giscus"></div>
	{{.Settings.AdditionalElemensBottom}}
</body>
</html>
`

var htmlIndexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	{{.Settings.AdditionalElementsTop}}
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Settings.Title}}</title>
    <link rel="stylesheet" href="style.css">
	<link rel="icon" type="image/x-icon" href="favicon.ico">
	<script src="https://cdn.jsdelivr.net/npm/fuse.js@7.0.0"></script>
	<script src="search.js"></script>
</head>
<body>
	<header>
		<h1>
			{{.Settings.Title}}
			<a href="rss.xml" target="_blank" title="RSS Feed">
			RSS
			</a>
		</h1>
		<input type="text" id="search-input" placeholder="Search... (supports Unix-like search commands)">
    	<ul id="search-results"></ul>
		<nav>
		{{range .PageList}}
			<a href="{{.LinkToSelf}}"  {{if $.Settings.OpenInNewTab}}target="_blank"{{end}}>{{.Title}}</a>
		{{end}}
		</nav>
		<h2>{{.Settings.Description}}</h2>
		<div id="buttons"></div>
        <aside></aside>
    </header>
	{{ $dateFormat := .Settings.DateFormat}}
    {{range .ArticleList}}
        <div class="detail">
            <div class="headline">
                <a href="{{.LinkToSelf}}" {{if $.Settings.OpenInNewTab}}target="_blank"{{end}}>
                    <h2>{{.Title}}</h2>
                </a>
                <div class="info">
                    <div class="tags">
                        {{range .Tags}}
                            <button class="on">{{.}}</button>
                        {{end}}
                    </div>
                    <h4 class="date">⋆ {{.Created.Format $dateFormat}}</h4>
                    <h4 class="date">♰ {{.Updated.Format $dateFormat}}</h4>
                </div>
            </div>
			{{if .CoverImagePath}}
				<img src="{{.CoverImagePath}}" alt="{{.Title}}">
			{{end}}
            <p class="description">{{.Description}}</p>
        </div>
    {{end}}
    <script src="script.js" async defer></script>
    {{.Settings.AdditionalElemensBottom}}
</body>
</html>
`

const rssTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
	<channel>
		<title>{{ .Settings.Title }}</title>
		<link>{{ .Settings.BaseUrl }}</link>
		<description>{{ .Settings.Description }}</description>
		<generator>Go Simple Blog Generator</generator>
		<lastBuildDate>{{ .BuildDate }}</lastBuildDate>
		<atom:link href="{{ .Settings.BaseUrl }}/rss.xml" rel="self" type="application/rss+xml" />
		{{- range .Articles }}
		<item>
			<title>{{ .Title | htmlEscape }}</title>
			<link>{{ buildArticleURL . $.Settings  }}</link>
			<guid>{{ buildArticleURL . $.Settings  }}</guid>
			<pubDate>{{ .Created | formatPubDate }}</pubDate>
			<description>{{ .Description | htmlEscape }}</description>
			<media:content 
				xmlns:media="http://search.yahoo.com/mrss/" 
				url="https://www.pudim.com.br/pudim.jpg" 
				medium="image" 
				type="image/jpeg" 
				width="150" 
				height="150"
			/>
		</item>
		{{- end }}
	</channel>
</rss>
`
