---
title: 01 Getting Started with DSBG
description: A comprehensive guide to using the Dead Simple Blog Generator.
created: 2024-07-04
updated: 2024-07-05 10:30:00
tags: [guide, tutorial, example]
coverImagePath: 01_dsbg_logo.webp
---

# DSBG (Dead Simple Blog Generator)

[![Go Version](https://img.shields.io/github/go-mod/go-version/tesserato/dsbg)](https://go.dev)
[![License](https://img.shields.io/github/license/tesserato/dsbg)](https://github.com/tesserato/dsbg/blob/main/LICENSE)
[![Release Status](https://img.shields.io/github/release/tesserato/dsbg)](https://github.com/tesserato/dsbg/releases)
<!-- [![Build Status](https://github.com/tesserato/dsbg/actions/workflows/go.yml/badge.svg)](https://github.com/tesserato/dsbg/actions/workflows/go.yml) -->

DSBG (Dead Simple Blog Generator) is a command-line tool that transforms a directory of Markdown and HTML files into a static website. It's designed for ease of use, allowing you to quickly create and deploy a personal blog, documentation site, or a simple website with minimal configuration.

## Installation

To get started with DSBG, you'll first need to install it on your system. If you have Go installed, you can use the following command in your terminal:

```bash
go install github.com/tesserato/dsbg@latest
```

This command will download and install the latest version of DSBG.

Alternatively, you can download a pre-built binary for your operating system from the [releases page](https://github.com/tesserato/dsbg/releases) and add it to your system's PATH. This will allow you to run the `dsbg` command from any directory.

After installation, verify that DSBG is working correctly by running the help command:

```bash
dsbg -h
```

This command will print the help message, displaying all available options and flags.

## Usage

Here are some examples to help you get started with DSBG:

### Quick Test Blog

To quickly create a test blog and see DSBG in action, run:

```bash
dsbg -title "My First Blog" -description "A simple test blog"
```

This command will generate a basic website with the title "My First Blog" and the description "A simple test blog".

### Creating a Markdown Template

To create a new Markdown file with a pre-filled frontmatter template, run:

```bash
dsbg -template -title "My Awesome Post" -description "A sample template"
```
This command generates a Markdown file named with the current date (YYYYMMDD) in the current directory. You can use this as a starting point to fill with content. If no title is given, the file will be named with the date.

### Building a Blog with a Theme and Watch Mode

To build a blog using a predefined theme (dark) and enable watch mode (which automatically rebuilds the site on changes) use:

```bash
dsbg -title "My Awesome Blog" -description "My blog description" -watch -style dark
```

This command generates a blog with the "dark" theme and watches for changes in the content folder, rebuilding the website on each change. A local server will also start, to serve the blog on your browser.

### Building a Blog with Custom CSS

To use a custom CSS file to style your blog use:

```bash
dsbg -title "My Awesome Blog" -description "My blog description" -watch -css-path assets/style-colorful.css
```

This command creates a blog that is watched for changes, while using a custom CSS file for styling. Make sure that the path given with `-css-path` is the correct path to your CSS file.

### Including Custom HTML Elements

To include custom HTML elements (like analytics code, headers, or footers) in your blog use:

```bash
dsbg -title "My Awesome Blog" -description "My blog description" -watch -elements-top elements-top.html -elements-bottom elements-bottom.html
```

This command adds the contents of `elements-top.html` at the top of each generated page, and the content of `elements-bottom.html` at the bottom of each generated page. The files given to these flags must be basic HTML files.

### Special "PAGE" Tag Usage

For stand-alone pages or custom directories you want to copy, add the tag `PAGE` to your Markdown or HTML file's frontmatter, under the tags field. Then, the program will copy the directory and all its contents to the output folder. This is especially useful for HTML pages that have their own custom CSS and images, or for special documentation pages.

## Features

Here's a breakdown of the core features of DSBG:

*   **Markdown & HTML Support:**  Write your content using Markdown for ease of use or basic HTML for more complex layouts. Mix and match these formats freely within the same site.
*   **Frontmatter:** Use YAML frontmatter in Markdown files to define metadata like the article's `title`, `description`, `created` and `updated` dates, `tags` for categorization, and `coverImagePath` to add a cover to each post.
*   **Tag Filtering:**  DSBG automatically generates tag-based filters on the main index page for easy navigation, allowing your visitors to find related content quickly.
*   **Theming:**  Choose between pre-defined themes (`default`, `dark`, `colorful`) or use your own custom CSS file to style your website the way you want.
*  **RSS Feed Generation:** DSBG automatically generates an RSS feed (`rss.xml`) for your blog, allowing your readers to subscribe to your content updates.
*   **Client-Side Search:** A simple client-side search powered by Fuse.js is included, enabling users to easily search for content within your website.
*   **Resource Handling:** DSBG automatically identifies and copies resources (images, scripts, etc.) linked in your Markdown or HTML content to the output directory.
*   **Automatic Date Handling:** DSBG extracts and displays date information from frontmatter fields, filenames, or file metadata, ensuring your content displays with accurate dates.
*   **Watch Mode:** When enabled, DSBG monitors your content directory for changes and automatically rebuilds the website and starts a local server, allowing for a smooth writing experience.
*   **"PAGE" Tag:** Use the `PAGE` tag in the frontmatter of a file to designate entire folders for copying into the public directory, which is useful for adding custom pages or resources.
*   **Open Links in New Tabs:** Choose to open external links in new browser tabs, providing a smoother user experience for your readers.

## Command-Line Arguments

The following flags are available when using the `dsbg` command-line tool:

| Flag                       | Description                                                                                                   | Default Value     |
|----------------------------|---------------------------------------------------------------------------------------------------------------|-------------------|
| `-title`          | The title of your blog or website.                                                                          | `Blog`            |
| `-base-url`        | The base URL for your website (e.g., `https://example.com`). Required for generating a correct RSS feed.         | none              |
| `-description`   | A brief description of your blog or website. Can be in Markdown format.                                       | `This is my blog` |
| `-input-path`     | The path to the directory containing your source content files (Markdown and HTML).                             | `content`         |
| `-output-path`    | The path to the directory where the generated website files will be saved.                                    | `public`          |
| `-date-format`     | The format for displaying dates on the website using Go's time formatting syntax (e.g., `2006-01-02`).         | `2006 01 02`      |
| `-index-name`     | The filename for your main index page.                                                                         | `index.html`      |
| `-css-path`        | The path to a custom CSS file to override default styles.                                                    | none              |
| `-js-path`         | The path to a custom JavaScript file to add functionality to your site.                                         | none              |
| `-favicon-path`    | The path to a custom favicon file (e.g., `.ico`) for your site.                                                 | none              |
| `-ignore-tags-from-paths`    | Disable extracting tags from directory names.                                                                | `false`           |
| `-keep-date-in-paths`       | Do not remove date patterns from the generated file paths.                                                      | `false`           |
| `-keep-date-in-titles`       | Do not remove date patterns from article titles.                                                              | `false`           |
| `-open-in-new-tab`          | Open external article links in a new browser tab.                                                             | `false`           |
| `-style`           | Choose a pre-defined theme: `default`, `dark`, or `colorful`.                                               | `default`         |
| `-elements-top`   | The path to an HTML file with elements to include at the top of each generated page (e.g., analytics scripts). | none              |
| `-elements-bottom`| The path to an HTML file with elements to include at the bottom of each generated page.                         | none              |
| `-help`                   | Display this help message and exit.                                                                           | `false`           |
| `-watch`                  | Enable watch mode, automatically rebuild the site on changes, and start a local server.                       | `false`           |
| `-template`              | Create a basic Markdown template file with frontmatter in the current directory.                             | `false`           |

## Caveats

*   The output directory (`public` by default) is completely reset (deleted and re-created) every time the program is run. Any files you have in this folder will be removed, so make sure to not store any important files there. This means that any custom JavaScript, CSS, or other files that are not in your input directory will be lost when you rebuild the blog.
*   Links to images or other resources must not have any whitespace characters.
*   If a HTML page has additional content in its own folder, then you must use the "PAGE" tag, to tell DSBG to copy the folder and all its contents to the output folder.
*   If DSBG cannot parse a file, it will show a warning and move on. If a required resource is missing, it will print a warning, and the resource will be missing on the generated website.

## Contributing

Contributions are welcome! If you find a bug, have a feature request, or would like to improve the documentation, feel free to submit a pull request or create an issue on [GitHub](https://github.com/tesserato/dsbg).
