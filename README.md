DSBG (Dead Simple Blog Generator) is a command-line tool that transforms a directory of Markdown and HTML files into a static website. It's designed for ease of use, allowing you to quickly create and deploy a personal blog or documentation site with minimal configuration.

# Installation

If you have Go installed run `go install github.com/tesserato/dsbg@latest`. Alternatively, you can download the appropriate binary from the [releases page](https://github.com/tesserato/dsbg/releases) and optionally add it to your PATH.

Then run `dsbg -h` to see the help message and make sure everything is set up correctly.

# Usage

Here are a few examples to get you started:

## Create a Markdown Template

```bash
dsbg -template -title "My Awesome Post" -description "A sample template"
```
This command generates a markdown file named with the current date (YYYY MM DD) in the current directory, ready to be filled with content. If you don't provide a title, the date will be the name of the generated file.

## Build a Blog with a Predefined Theme and Watch Mode

```bash
dsbg -title "My Awesome Blog" -description "My blog description" -watch -style dark
```

This command generates a blog with the "dark" theme and watches for changes in the content folder, rebuilding and serving the site on each change.

## Build a Blog with Custom CSS

```bash
dsbg -title "My Awesome Blog" -description "My blog description" -watch -css-path assets/style-colorful.css
```

This command creates a blog that is watched for changes, while using a custom CSS for styling.

## Including Custom HTML Elements

```bash
dsbg -title "My Awesome Blog" -description "My blog description" -watch -elements-top elements-top.html -elements-bottom elements-bottom.html
```

This command adds the contents of `elements-top.html` at the top of each page (useful for things like analytics tracking) and `elements-bottom.html` at the bottom of each page (useful for comments section or footers).

<!-- ### Example of a PAGE Tag
For stand-alone pages or custom directories to copy, you can add the tag `PAGE` to the markdown or HTML file. Then, the program will copy the directory and all its contents to the output folder. -->



# Features:

*   **Markdown & HTML Support:** Write content using Markdown or basic HTML. Mix and match as needed.
*   **Frontmatter:** Use YAML frontmatter in Markdown files to define metadata like title, description, dates, tags, and cover images.
*   **Tag Filtering:** Automatically generates tag-based filters on the main index page for easy content navigation.
*   **Theming:** Offers pre-defined themes (`default`, `dark`, `colorful`) and the option to use a custom CSS file.
*   **RSS Feed Generation:** Automatically creates an RSS feed (`rss.xml`) for your blog, making it easy for readers to subscribe.
*   **Client-Side Search:**  Includes a simple client-side search functionality powered by Fuse.js, allowing users to find content easily.
*   **Resource Handling:** Automatically copies images, scripts, and other linked resources from your content files to the output directory.
*   **Automatic Date Handling:** Extracts and displays date information from frontmatter, filenames, and file metadata.
*   **Watch Mode:**  Monitors your content directory for changes and automatically rebuilds the site. Includes a built-in local server for easy preview.
*  **"PAGE" Tag:** Special tag to designate pages by copying all the folder contents to the public folder.
*   **Open Links in New Tabs:** Option to open external article links in a new browser tab for a smoother user experience.


<!-- # Command-Line Arguments

*   `-title <string>`:  The title of your blog or website. (Default: `Blog`)
*   `-base-url <string>`:  The base URL for your website (e.g., `https://example.com`). Required for generating a correct RSS feed. (Default: none)
*   `-description <string>`: A brief description of your blog or website. (Default: `This is my blog`)
*   `-input-path <string>`: The path to the directory containing your source content files (Markdown and HTML). (Default: `content`)
*   `-output-path <string>`: The path to the directory where the generated website files will be saved. (Default: `public`)
*   `-date-format <string>`: The format for displaying dates on the website using Go's time formatting syntax (e.g.  `2006-01-02` , `02/01/2006`). (Default: `2006 01 02`)
*   `-index-name <string>`: The filename for your main index page. (Default: `index.html`)
*   `-css-path <string>`: The path to a custom CSS file to override default styles. (Default: none)
*   `-js-path <string>`: The path to a custom JavaScript file to add functionality to your site. (Default: none)
*  `-favicon-path <string>`: The path to a custom favicon file (e.g., `.ico`) for your site. (Default: none)
*   `-ignore-tags-from-paths`: Disable extracting tags from directory names. (Default: `false`)
*   `-keep-date-in-paths`: Do not remove date patterns from the generated file paths. (Default: `false`)
*   `-keep-date-in-titles`: Do not remove date patterns from article titles. (Default: `false`)
*   `-open-in-new-tab`: Open external article links in a new browser tab. (Default: `false`)
*   `-style <string>`: Choose a pre-defined theme: `default`, `dark`, or `colorful`. (Default: `default`)
*   `-elements-top <string>`:  The path to an HTML file with elements to include at the top of each generated page (e.g., analytics scripts, header). (Default: none)
*   `-elements-bottom <string>`:  The path to an HTML file with elements to include at the bottom of each generated page (e.g., comments section, footer). (Default: none)
*   `-help`: Display this help message and exit. (Default: `false`)
*   `-watch`:  Enable watch mode, automatically rebuild the site on changes, and start a local server. (Default: `false`)
*   `-template`:  Create a basic Markdown template file with frontmatter in the current directory. (Default: `false`) -->


# Caveats

*  The output directory (`public` by default) is reset every time the program is run.
*  no whitespace in link to image; 
*  pages with additional content in its on folder

# Contributing

Your contributions are welcome! Feel free to submit pull requests or create issues for any bugs or new features you'd like to see.

