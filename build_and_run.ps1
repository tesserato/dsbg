$exec_name = "dsbg.exe"
$front_matter = "---
title: README - Getting Started with DSBG
description: How to install and use Dead Simple Blog Generator.
created: 2025 01 03
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

# ./dsbg.exe -template -title "My Awesome Post from template" -description "A sample template with a very long text to test if it wraps correctly. Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua." -output-path "sample_content"

$description = @'
A Simple, Open-Source Tool to Create Your Static Blog and Broadcast Your Content.

# TLDR

`go install github.com/tesserato/dsbg@latest` or download a [pre-built binary](https://github.com/tesserato/dsbg/releases)

`dsbg -h` for usage instructions

Check the Readme [here](https://tesserato.github.io/dsbg/01readme/index.html) or at the Github [repo](https://github.com/tesserato/dsbg) for more details

This is a sample blog created with DSBG from the source at [github.com/tesserato/dsbg](https://github.com/tesserato/dsbg/tree/main/sample_content)

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
    -input-path "sample_content" `
    -output-path "docs" `
    -base-url "https://tesserato.github.io/dsbg/" `
    -elements-top "analytics.html" `
    -elements-bottom "giscus.html" `
    -style "default" `
    -x-handle "tesserato" `
    -bluesky-handle "tesserato" `
    -threads-handle "tesserato" `
    -mastodon-handle "tesserato" `
    -sort "reverse-date-created"
    
    