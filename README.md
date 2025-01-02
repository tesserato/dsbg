# Description

Given a directory of markdown and html files, it generates a static website.

It accepts a directory as input, a name for the output directory, and various configuration options.

The generated website features:

-   **Markdown and HTML support:** Create content using simple Markdown or basic HTML.
-   **Frontmatter:** Use frontmatter to define metadata like title, description, dates, tags, and cover images.
-   **Tag Filtering:** Automatically generate tag buttons for filtering articles on the index page.
-   **Theming:** Supports predefined themes (default, dark, colorful) and custom CSS.
-   **RSS Feed Generation:** Automatically creates an RSS feed for your blog.
-   **Client-side search:** Simple client-side searching using Fuse.js
-   **Automatic resource handling:** Copy images, scripts, etc., linked in articles to the output directory.
-   **Automatic date handling:** Extracts and displays date information from frontmatter, filenames, and file metadata.
-   **Watch Mode:** Automatically rebuild the site and serve it with a local server on file changes.
-   **"PAGE" Tag:** Special tag to handle stand-alone pages by copying all directory content.
-   **Open Links in New Tabs:** Option to open external article links in a new tab.

# Principles

-   **Out of the box:** Ready to go with minimal setup.
-   **DRY(ish):** Reduces redundancy in common blog setups.
-   **Explicit:** Clear command-line arguments and structure.

# Command-line Arguments

-   `-title <string>`: The title of the blog. (Default: `Blog`)
-   `-base-url <string>`: The base URL for the blog (e.g., https://example.com) used when generating the RSS feed and must be set if you intend to use the rss feed. (Default: none)
-   `-description <string>`: A short description of the blog. (Default: `This is my blog`)
-   `-input-path <string>`: Path to the directory containing source files (Markdown or HTML). (Default: `content`)
-   `-output-path <string>`: Path to the directory where the generated website files will be saved. (Default: `public`)
-   `-date-format <string>`: Format for displaying dates (Go time format). (Default: `2006 01 02`)
-   `-index-name <string>`: Filename for the main index page. (Default: `index.html`)
-   `-css-path <string>`: Path to a custom CSS file. (Default: none)
-   `-js-path <string>`: Path to a custom JavaScript file. (Default: none)
-   `-favicon-path <string>`: Path to a custom favicon file (e.g., `.ico`). (Default: none)
-   `-ignore-tags-from-paths`: Disable extracting tags from directory names. (Default: `false`)
-   `-keep-date-in-paths`: Do not remove date patterns from generated file paths. (Default: `false`)
-   `-keep-date-in-titles`: Do not remove date patterns from article titles. (Default: `false`)
-   `-open-in-new-tab`: Open article links in a new browser tab. (Default: `false`)
-   `-style <string>`: Predefined style to use (`default`, `dark`, `colorful`). (Default: `default`)
-   `-elements-top <string>`: Path to an HTML file with elements to include at the top of each page (e.g., analytics scripts). (Default: none)
-   `-elements-bottom <string>`: Path to an HTML file with elements to include at the bottom of each page (e.g., comments section). (Default: none)
-   `-help`: Show this help message and exit. (Default: `false`)
-   `-watch`: Watch for changes in the input directory and rebuild automatically. Also creates a server to serve the website. (Default: `false`)
-   `-template`: Create a basic Markdown template file with frontmatter. (Default: `false`)

# Example Usage

## Create a Markdown Template
```bash
./dsbg.exe -template -title "My Awesome Post" -description "A sample template"
```

This will generate a markdown file in the current directory.

## Create a Blog using predefined styles and a local server
./dsbg.exe -title "My Awesome Blog" -description "My blog description" -watch -style dark


## Create a Blog using custom CSS
./dsbg.exe -title "My Awesome Blog" -description "My blog description" -watch -css-path assets/style-colorful.css


## Options for custom HTML elements
./dsbg.exe -title "My Awesome Blog" -description "My blog description" -watch -elements-top elements-top.html -elements-bottom elements-bottom.html


where elements-top.html and elements-bottom.html are custom files with html elements that will be added to the generated website.
This is used to add things such as analitics scripts and comments section.


# Caveats

Output directory gets reset every time the program is run.

Ensure correct paths are given for custom css, javascript and favicon files.

Ensure to include base-url if you intend to use the rss.xml generated.

# Contributing

Feel free to submit pull requests or create issues for any bugs or new features you would like to see implemented.

