package controllers

import (
	u "cig-exchange-libs"
	"fmt"
	"net/http"
	"text/template"
)

// defaultTemplates are included every time a template is rendered.
var defaultTemplates = []string{"/user_data/templates/base.html", "/user_data/templates/info.html"}

// Ping handles the ping test request
var Ping = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "Pong!")
	u.Respond(w, resp)
}

// Login renders the login/registration page.
var Login = func(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login.html", nil)
}

// renderTemplate renders the template to the ResponseWriter
func renderTemplate(w http.ResponseWriter, f string, data interface{}) {
	t, err := template.ParseFiles(append(defaultTemplates, fmt.Sprintf("/user_data/templates/%s", f))...)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	t.ExecuteTemplate(w, "base", data)
}
