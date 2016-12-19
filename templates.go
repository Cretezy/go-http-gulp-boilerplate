package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/context"
)

type TemplateManager struct {
	templates map[string]*template.Template
}

func NewTemplateManager() *TemplateManager {
	fmap := template.FuncMap{
		"IsCurrentView": func(current string, test string) string {
			if current == test || strings.Split(current, ".")[0] == test {
				return "active"
			} else {
				return ""
			}
		},
	}

	templates := make(map[string]*template.Template)

	templatesDir := "views"

	views, err := AssetDir(templatesDir)
	if err != nil {
		log.Fatal(err)
	}

	viewsFiltered := make([]string, 0)
	for _, view := range views {
		if strings.HasSuffix(view, ".tmpl") {
			viewsFiltered = append(viewsFiltered, view)
		}
	}

	views = viewsFiltered

	layouts, err := AssetDir(templatesDir + "/layouts")
	if err != nil {
		log.Fatal(err)
	}

	layoutsFiltered := make([]string, 0)
	for _, layout := range layouts {
		layoutsFiltered = append(layoutsFiltered, "layouts/"+layout)
	}

	layouts = layoutsFiltered

	// Generate our templates map from our layouts/ and includes/ directories
	for _, view := range views {
		files := append(layouts, view)
		tmpl := template.New(filepath.Base(view)).Funcs(fmap)
		for _, file := range files {
			content, err := Asset("views/" + file)
			if err != nil {
				panic("cannot find template " + file)
			}
			template.Must(tmpl.Parse(string(content)))
		}
		templates[filepath.Base(view)] = tmpl
	}

	return &TemplateManager{
		templates,
	}
}

// renderTemplate is a wrapper around template.ExecuteTemplate.
func (templateManager TemplateManager) RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	// Ensure the template exists in the map.
	tmpl, ok := templateManager.templates[name]
	if !ok {
		return fmt.Errorf("The template %s does not exist.", name)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, "base", data)
}

func (templateManager TemplateManager) RenderView(w http.ResponseWriter, r *http.Request, view string, errors []string, data interface{}) error {
	user := new(User)
	var loggedIn bool
	if context.Get(r, "user") != nil {
		user = context.Get(r, "user").(*User)
		loggedIn = true
	}

	viewData := View{
		Errors:      errors,
		Data:        data,
		User:        user,
		LoggedIn:    loggedIn,
		CurrentView: view,
	}

	return templateManager.RenderTemplate(w, view+".tmpl", viewData)
}

type View struct {
	Errors      []string
	LoggedIn    bool
	User        *User
	CurrentView string
	Data        interface{}
}
