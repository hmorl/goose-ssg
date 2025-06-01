# GOose-ssg

_Honkin' simple and very silly static site generator written in Go_

No markdown 🫨, no YAML (🥂), certainly no server-side, no databases, no-HONNNNKKK 🪿

:warning: This tool was written in a couple of hours. It might not be what you want... :warning:

## What IS it then? 🪿

GOose-ssg is a static site generator designed to ease the creation of a basic HTML website, i.e. a handful of pages which might share a common layout or common elements (such as a nav bar). This could be a quick project site, a landing page or piece of web art. Anything small.

**INCOMING MESSAGE FROM THE GOOSE**

🪿: "Not everything needs to be a full-blown CMS or markdown blog, honey"

It leans on Go's [HTML template package](https://pkg.go.dev/html/template) & templating system.

## Features

- Site generator which processes a base layout HTML template, and "content" HTML pages (which are themselves just template definitions of the content, made use of by the base template... see example below)
- Localhost file server for previewing the site
- File watcher with hot reload for quick iteration
- Automatic and opinionated static routing which turns everything into a hierarchy of `index.html`s. For a handful of pages this means you don't have to make as many subdirectories compared to a raw HTML site, if all you're after are some pages with clean URLs like `mysite.com/hello`, `mysite.com/hello/world`.
- Bundling of static content (sounds fancy but it just dumps `static/` into the root of your site)
- `map` function for passing around key-value pairs between templates
- `ThisPage` variable passed into the base template, for doing things like highlighting the current page on your cute nav bar... e.g.

```
{{define "nav-link"}}
<a href="{{.Url}}" {{if eq .Root.ThisPage .Url }}class="nav-active" {{end}}>{{.Text}}</a>
{{end}}
```

## Example

Given a hierarchy of pages and some simple templates:

```
pages/
    index.html
    about.html
    collection.html
    collection/
        coins.html
        stamps.html
templates/
    base.html
    nav-bar.html
```

Where `base.html` might look like...

```
{{define "base"}}
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <link rel="stylesheet" href="/style.css" />
    <title>Site title</title>
  </head>

  <body>
    <header>Header</header>
    <main>{{block "main".}}{{end}}</main>
    <footer>Footer</footer>
  </body>
</html>
{{end}}
```

and `index.html` might look like...

```
{{define "main"}}
<h1>Hello</h1>
<p>
Welcome to the home page!
</p>
{{end}}
```

GOose will take the pages, process the templating, and generate the following site:

```
dist/
    index.html
    about/
        index.html
    collection/
        index.html
        coins/
            index.html
        stamps/
            index.html
```

`mysite.com/collection` will go to the content which was defined in `collection.html` and `mysite.com/collection/coin` will go to the coin collection content (yawn).

As well as building the site, GOose-ssg can also locally serve the site and watch for changes (with hot reload), by passing in `--serve`

## Real example

See `test/testdata` for another example. Generate it either manually with the compiled goose-ssg executable or by running the integration tests with `make test-cli`, and see the generated site in `test/dist`

## Building

Prerequisites:

- a recent version of [Go](https://go.dev/), available in your path.
- (optional) GNU Make for using the Makefile. Otherwise, read the commands in the Makefile and run them directly.

After cloning the repo and navigating to it from a terminal, simply run `make` or `make build`.

## Current limitations

- Only supports a single base template/layout, called `base`. Could support a simple config file for mapping pages to different base templates in the future!
- Server is fixed to `localhost:3000`
- Loads of other limitations which are intentional. 🪿.

## See also

- [Hugo](https://gohugo.io/) - a *far superior* SSG. But it has far _more stuff_.
