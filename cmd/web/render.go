package main

import (
	"fmt"
	"github.com/kennnyz/concurrencyGO/final-project/data"
	"html/template"
	"net/http"
	"time"
)

var pathTemplates = "./cmd/web/templates"

type TemplateData struct {
	StringMap     map[string]string
	IntMap        map[string]string
	FloatMap      map[string]float64
	Data          map[string]any
	Flash         string
	Warning       string
	Error         string
	Authenticated bool
	Now           time.Time
	User          *data.User
}

func (app *Config) render(w http.ResponseWriter, r *http.Request, t string, templateData *TemplateData) {
	partials := []string{
		fmt.Sprintf("%s/base.layout.gohtml", pathTemplates),
		fmt.Sprintf("%s/header.partial.gohtml", pathTemplates),
		fmt.Sprintf("%s/navbar.partial.gohtml", pathTemplates),
		fmt.Sprintf("%s/footer.partial.gohtml", pathTemplates),
		fmt.Sprintf("%s/alerts.partial.gohtml", pathTemplates),
	}

	var templateSlice []string
	templateSlice = append(templateSlice, fmt.Sprintf("%s/%s", pathTemplates, t))

	for _, x := range partials {
		templateSlice = append(templateSlice, x)
	}

	if templateData == nil {
		templateData = &TemplateData{}
	}

	tmpl, err := template.ParseFiles(templateSlice...)
	if err != nil {
		app.ErrorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, app.AddDefaultData(templateData, r)); err != nil {
		app.ErrorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *Config) AddDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")
	if app.IsAuthenticated(r) {
		td.Authenticated = true
		user, ok := app.Session.Get(r.Context(), "user").(data.User)
		if !ok {
			app.ErrorLog.Println("Can't get user from session")
		} else {
			td.User = &user
		}

	}
	td.Now = time.Now()

	return td
}

func (app *Config) IsAuthenticated(r *http.Request) bool {
	return app.Session.Exists(r.Context(), "userID")
}
