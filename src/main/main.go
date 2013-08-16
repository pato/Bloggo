package main

import (
	"fmt"
	"text/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"github.com/russross/blackfriday"
)

type Page struct {
	Title string
	Body  []byte
	Render []byte
}

type HomePage struct{
	Pages []string
}

func (p *Page) save() error {
	filename := p.Title + ".page"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".page"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	render := blackfriday.MarkdownCommon(body)
	return &Page{Title: title, Body: body, Render:render}, nil
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, ".") {
		fmt.Printf("Served: %s\n", r.URL.Path)
		chttp.ServeHTTP(w, r)
	} else {
		// fmt.Fprintf(w, "<html>\n")
		// fmt.Fprintf(w, "<head>\n<style>\nbody{background-color:black;}\n.dir{color:#2BFF00; text-decoration:none;}\n.back{color:#FF0000; text-decoration:none;}\n.file{color:#FFFF00; text-decoration:none;}</style>\n</head>\n")

		dir, _ := ioutil.ReadDir("." + r.URL.Path[1:])

		var pages []string

		for _, entry := range dir {
			if strings.Contains(entry.Name(),".page") {
				pages = append(pages, strings.TrimSuffix(entry.Name(),".page"))
			}
		}
		// for _, r := range pages {
		// 	name := strings.TrimSuffix(r,".page")
		// 	fmt.Fprintf(w, "--- <a class='file' href='/view/%s'>%s</a><br>\n", name, r)
		// }
		//fmt.Fprintf(w, "</html>")
		renderHome(w, HomePage{pages})
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html", "home.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderHome(w http.ResponseWriter, h HomePage) {
	err := templates.ExecuteTemplate(w, "home.html", h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const lenPath = len("/view/")

var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path[lenPath:]
		if !titleValidator.MatchString(title) {
			http.NotFound(w, r)
			return
		}
		fn(w, r, title)
	}
}

var chttp = http.NewServeMux()
func main() {
	chttp.Handle("/", http.FileServer(http.Dir(".")))
	fmt.Printf("Starting...\n")
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	fmt.Printf("Listening on 8080...\n")
	http.ListenAndServe(":8080", nil)
}
