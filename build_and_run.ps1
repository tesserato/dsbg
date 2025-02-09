$exec_name = "dsbg.exe"
$front_matter = "---
title: 01 Getting Started with DSBG
description: How to install and use Dead Simple Blog Generator.
created: 2025 01 04
updated: 2025 01 04
tags: [guide, tutorial, example]
coverImagePath: 01_dsbg_logo.webp
---"

if (Test-Path $exec_name) {
    Remove-Item $exec_name
}

# go build -ldflags="-s -w" .
go get -u
go get -u ./...
go mod tidy
go build .

Remove-Item "docs/*" -Recurse -Force

$content = Get-Content README.md -Encoding UTF8 -Raw
$content = $front_matter + "`n`n" + $content
Set-Content sample_content/01_readme.md $content

magick -density 376 -background none "logo.svg" "sample_content/01_dsbg_logo.webp"
magick -background none "sample_content/01_dsbg_logo.webp" -fill red -opaque black -blur 0x1  -crop 167x167+0+0  "assets/favicon.ico"

./dsbg.exe -template -title "My Awesome Post" -description "A sample template" -output-path "sample_content"

$description = @'
Welcome to the Dead Simple Blog Generator blog.

A new static site generator that's truly simple. Built with Go for speed & efficiency.

- Easy installation: Download a [pre-built binary](https://github.com/tesserato/dsbg/releases) or use `go install github.com/tesserato/dsbg@latest`,

- Support for Markdown & HTML source files, 

- Tag generation and filtering, 

- Client-side fuzzy search,

- RSS feed generation,

- Watch mode with automatic re-rendering,

- 3 different themes, with the ability to add your own via custom CSS.

- Share buttons

- Easy analytics, comments, and more.

This is a sample blog created with DSBG from the source at [github.com/tesserato/dsbg](https://github.com/tesserato/dsbg)

[![Release Status](https://img.shields.io/github/release/tesserato/dsbg)](https://github.com/tesserato/dsbg/releases)

[![License](https://img.shields.io/github/license/tesserato/dsbg)](https://github.com/tesserato/dsbg/blob/main/LICENSE)

'@

# [![Build Status](https://github.com/tesserato/dsbg/actions/workflows/go.yml/badge.svg)](https://github.com/tesserato/dsbg/actions/workflows/go.yml)

# [![Go Version](https://img.shields.io/github/go-mod/go-version/tesserato/dsbg)](https://go.dev/)

start chrome http://localhost:666/index.html

./dsbg.exe `
    -title "Dead Simple Blog Generator" `
    -description "$description" `
    -watch `
    -open-in-new-tab `
    -input-path "sample_content" `
    -output-path "docs" `
    -base-url "https://tesserato.github.io/dsbg/" `
    -elements-top "analytics.html" `
    -style "default" `
    -x-handle "tesserato" `
    -bluesky-handle "tesserato" `
    -threads-handle "tesserato" `
    -mastodon-handle "tesserato" `
    -sort "date-updated"
    
    