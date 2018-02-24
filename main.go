package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// check for required layout and home files
	if _, err := os.Stat("content/home.gohtml"); os.IsNotExist(err) {
		log.Fatalln("home.gohtml does not exist")
	}
	if _, err := os.Stat("content/layout.gohtml"); os.IsNotExist(err) {
		log.Fatalln("layout.gohtml does not exist")
	}

	// if not dev mode, parse templates during program load (otherwise, parse at each request)
	var dev = len(os.Args) > 1 && strings.ToUpper(os.Args[1]) == "DEV"
	var pages Pages //TODO: Race condition exists in DEV mode; eliminate this with a mutex
	if !dev {
		var err error
		pages, err = parsePages(dev)
		if err != nil {
			log.Fatalln("problem during production startup:", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if dev {
			var err error
			pages, err = parsePages(dev)
			if err != nil {
				log.Println("problem during development request:", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		app := App{
			Brand: getBrand(),
			Nav:   buildNav(pages),
		}
		switch {
		case r.URL.Path == "/":
			pages["home"].Execute(w, app)
		case isContentFile(r.URL):
			fs := http.FileServer(http.Dir("content"))
			fs.ServeHTTP(w, r)
		default:
			if val, ok := pages[lastPart(r.URL)]; ok {
				val.Execute(w, app)
				return
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
	})
	srv := &http.Server{
		Addr:    ":3033",
		Handler: mux,
		//time from when the connection is accepted to when the request body is fully read
		ReadTimeout: 5 * time.Second,
		//time from the end of the request header read to the end of the response write
		WriteTimeout: 10 * time.Second,
	}
	fmt.Println("Now serving on port 3033")
	log.Println(srv.ListenAndServe())
}

func lastPart(url *url.URL) string {
	pathParts := strings.Split(url.Path, `/`)
	return pathParts[len(pathParts)-1]
}

func isContentFile(url *url.URL) bool {
	part := lastPart(url)
	// Does last portion of URL contain a '.'? Make sure it isn't gohtml?
	return strings.Contains(part, ".") && !strings.Contains(part, ".gohtml")
}

func parsePages(isDev bool) (Pages, error) {
	pages := make(map[string]*template.Template)
	pageFiles, err := filepath.Glob("content/*.gohtml")
	if err != nil {
		return pages, fmt.Errorf("could not parse pages: %s", err)
	}
	for _, f := range pageFiles {
		if f != filepath.Join("content", "layout.gohtml") {
			tmpl, err := template.ParseFiles("content/layout.gohtml", f)
			if err != nil {
				return pages, fmt.Errorf("could not parse template: %s", err)
			}
			pageName := strings.Split(f, ".")[0]
			pages[filepath.Base(pageName)] = tmpl
		}
	}
	return pages, nil
}

// Pages is a map of kebabcase names to templates
type Pages map[string]*template.Template

// App specifies Brand and Nav
type App struct {
	Brand string
	Nav   Nav
}

// NavItem specifies the text and URL for a link on the top nav bar
type NavItem struct {
	Text string
	URL  template.URL
}

// Nav is a slice of NavItem
type Nav []NavItem

func buildNav(pages Pages) Nav {
	menu := make([]NavItem, 0, len(pages))
	// add Home nav item first
	homeItem := NavItem{
		Text: "Home",
		URL:  template.URL("/"),
	}
	menu = append(menu, homeItem)
	for t := range pages {
		if t != "home" {
			item := NavItem{
				Text: kebabToTitle(t),
				URL:  template.URL("/" + t),
			}
			menu = append(menu, item)
		}
	}
	return menu
}

func kebabToTitle(kebab string) string {
	kebab = strings.Replace(kebab, "-", " ", -1)
	return strings.Title(kebab)
}

func getBrand() string {
	wd, _ := os.Getwd()
	parts := strings.Split(wd, string(os.PathSeparator))
	dir := parts[len(parts)-1]
	return strings.ToTitle(dir)
}
