---
title: 01 Getting Started with DSBG
description: A comprehensive guide to using the Dead Simple Blog Generator.
created: 2024-07-04
updated: 2024-07-05 10:30:00
tags: [guide, tutorial, example]
coverImagePath: 01_dsbg_logo.webp
---

# Welcome to the Dead Simple Blog Generator!

This post will guide you through the basics of using DSBG and demonstrate its key features.

## Core Features

DSBG is designed to be straightforward and easy to use. Here are some of its main capabilities:

*   **Markdown and HTML Support:** Write your content using Markdown for quick formatting or use standard HTML for more advanced customization. Mix them freely!
*   **Frontmatter:** Use YAML frontmatter (like the metadata at the top of this file) to define titles, descriptions, dates, tags, and cover images.
*   **Tag Filtering:** DSBG automatically creates filter buttons based on your tags, making it easy to categorize and browse content on your index page.
*   **Theming:** Comes with basic default themes and the option to use a custom CSS file for full control over styling.
*   **RSS Feed:**  DSBG generates a valid RSS feed (`rss.xml`) so your readers can subscribe using their favorite reader.
*   **Search:**  Uses a simple client-side search, allowing users to quickly find relevant content.

## Markdown Content Examples

Here's a mix of common Markdown elements to show you how they look:

### Lists

- An unordered list
    - Nested list item
- Another list item

1.  An ordered list
2.  Second item
    1.  Nested ordered list

### Code

You can include code blocks with syntax highlighting:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, DSBG!")
}
```

### Images

Images are automatically copied to the output directory:

![A Butterfly](01_butterfly.webp)

### Text Formatting

Here's some **bold** and *italic* text, and some `inline code`.

### Links and Quotes

Check out this [example link](https://example.com).

> This is a block quote. Useful for citing resources.

### More Text

Here is a normal paragraph, with a few sentences. DSBG will render this correctly. It should provide an example of normal usage.

## Conclusion

This post should give you a good overview of the features that DSBG offers and how to use them. Feel free to experiment with this content and check the generated HTML.

If you find any bugs, or have a feature you would like to see implemented, create an issue or a pull request.