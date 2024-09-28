package parse

var htmlArticleTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	{{.Settings.AdditionalElementsTop}}
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="generator" content="ZBSCM">
    <link rel="stylesheet" href="{{.Lks.ToCss}}">
    <link rel="icon" type="image/x-icon" href="../favicon.ico">
    <title>{{.Art.Title}}</title>
</head>

<body>
    <header>
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

var HtmlIndexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	{{.Settings.AdditionalElementsTop}}
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Settings.Title}}</title>
    <link rel="stylesheet" href="style.css">
	<link rel="icon" type="image/x-icon" href="favicon.ico">
</head>
<body>
	<header>
		<h1>{{.Settings.Title}}</h1>
		<nav>
		{{range .PageList}}
			<a href="{{.LinkToSelf}}">{{.Title}}</a>
		{{end}}
		</nav>
		<div id="buttons"></div>
        <aside></aside>
    </header>
	{{ $dateFormat := .Settings.DateFormat}}
    {{range .ArticleList}}
        <div class="detail">
            <div class="headline">
                <a href="{{.LinkToSelf}}">
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
            <p class="description">{{.Description}}</p>
        </div>
    {{end}}
    <script src="script.js" async defer></script>
    {{.Settings.AdditionalElemensBottom}}
</body>
</html>
`
