// Copyright 2014 Quincy Bowers.  All rights reserved.

package main

import (
	"bufio"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// Page represents a single page in the wiki.
type Page struct {
	Title   string
	Query   string
	Matches []string
	Min     int
	Max     int
	Count   int
}

func createPage(pattern string, min, max int) *Page {
	matches, count := getMatches(pattern, min, max)
	page := Page{
		Title:   "wordsearcher",
		Query:   pattern,
		Matches: matches,
		Min:     min,
		Max:     max,
		Count:   count}

	return &page
}

func getMatches(query string, min, max int) (matches []string, count int) {
	if query == "" {
		query = ".*"
	}

	pattern := regexp.MustCompile(query)

	for _, word := range words {
		if min > 0 && len(word) < min {
			continue
		}
		if max > 0 && len(word) > max {
			continue
		}
		if pattern.MatchString(word) {
			matches = append(matches, word)
		}
	}

	return matches, len(matches)
}

// viewHandler prepares the page to be rendered by passing it through the
// markdown and wikiMarkup filters.
func searchHandler(w http.ResponseWriter, r *http.Request, params url.Values) {
	var query string
	var min int
	var max int
	var err error

	if _, exists := params["query"]; !exists {
		query = ""
	} else {
		query = params["query"][0]
	}
	if _, exists := params["min"]; !exists {
		min = 0
	} else {
		min, err = strconv.Atoi(params["min"][0])
		if err != nil {
			panic(err)
		}
	}
	if _, exists := params["max"]; !exists {
		max = 0
	} else {
		max, err = strconv.Atoi(params["max"][0])
		if err != nil {
			panic(err)
		}
	}

	p := createPage(query, min, max)
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

func makeHandler(fn func(http.ResponseWriter, *http.Request, url.Values)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, r.URL.Query())
	}
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

var words []string

func init() {
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}

	dictionary := path.Join(currentUser.HomeDir, "words")
	dict, err := readLines(dictionary)
	if err != nil {
		panic(err)
	}

	var wordMap map[string]bool = make(map[string]bool)

	for _, line := range dict {
		word := strings.Trim(line, "\n")

		if strings.IndexAny(word, "'ABCDEFGHIJKLMNOPQRSTUVWXYZ") >= 0 {
			continue
		}

		if _, exists := wordMap[word]; !exists {
			wordMap[word] = true
			words = append(words, word)
		}
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
	http.HandleFunc("/search", makeHandler(searchHandler))
	http.ListenAndServe(server, nil)
}
