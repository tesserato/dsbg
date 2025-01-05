$exec_name = "dsbg.exe"
$front_matter = "---
title: 01 Getting Started with DSBG
description: A comprehensive guide to using the Dead Simple Blog Generator.
created: 2024-07-04
updated: 2024-07-05 10:30:00
tags: [guide, tutorial, example]
coverImagePath: 01_dsbg_logo.webp
---"

if (Test-Path $exec_name) {
    Remove-Item $exec_name
}

# go build -ldflags="-s -w" .
go mod tidy
go build .

Remove-Item "docs/*" -Recurse -Force

$content = Get-Content README.md -Encoding UTF8 -Raw
$content = $front_matter + "`n`n" + $content
Set-Content sample_content/01_readme.md $content

magick -density 376 -background none "logo.svg" "sample_content/01_dsbg_logo.webp"
magick -background none "sample_content/01_dsbg_logo.webp" -fill red -opaque black -blur 0x1  -crop 167x167+0+0  "assets/favicon.ico"

./dsbg.exe -template -title "My Awesome Post" -description "A sample template" -output-path "sample_content"

start chrome http://localhost:666/index.html

./dsbg.exe -title "Dead Simple Blog Generator" `
    -description "Welcome to the DSBG (Dead Simple Blog Generator) blog. 
    This is a sample blog created with DSBG from the source at [github.com/tesserato/dsbg](https://github.com/tesserato/dsbg)" `
    -watch -open-in-new-tab -css-path "assets/style.css" -input-path "sample_content" -output-path "docs" -base-url "https://tesserato.github.io/dsbg/"