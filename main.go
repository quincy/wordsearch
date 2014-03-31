// Copyright 2014 Quincy Bowers.  All rights reserved.

package main

import (
	"html/template"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Page represents a single page in the wiki.
type Page struct {
	Title   string
	Matches []template.HTML
}

func createPage(pattern string) *Page {
	return new(Page)
}

// viewHandler prepares the page to be rendered by passing it through the
// markdown and wikiMarkup filters.
func searchHandler(w http.ResponseWriter, r *http.Request, pattern string) {
	p := createPage(pattern)
	renderTemplate(w, "view", p)
}

// Parse the templates.
var templateDir string = "templates"
var templateFiles []string = []string{filepath.Join(templateDir, "view.html")}

var templates = template.Must(template.ParseFiles(templateFiles...))

// renderTemplate takes the renders the html for the given template.
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, r.URL.Path)
	}
}

func main() {
	var server = "localhost:8080"

	// open the default browser to the view/Home endpoint.
	var browser *exec.Cmd
	var url string = "http://" + server + "/"

	switch runtime.GOOS {
	case "windows":
		browser = exec.Command(`C:\Windows\System32\rundll32.exe`, "url.dll,FileProtocolHandler", url)
	case "darwin":
		browser = exec.Command("open", url)
	default:
		browser = exec.Command("xdg-open", url)
	}
	if err := browser.Start(); err != nil {
		panic(err)
	}

	// register the handlers and start the server.
	http.HandleFunc("/", makeHandler(searchHandler))
	http.ListenAndServe(server, nil)
}
