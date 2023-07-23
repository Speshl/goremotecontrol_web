package server

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
)

type IndexPageData struct {
	HeaderHTML template.HTML
	NavHTML    template.HTML
	MainHTML   template.HTML
	FooterHTML template.HTML
}

type LoginFormData struct {
	IsLoggedIn bool
	Username   string
	Rank       string
}

func (s *Server) buildIndex(w http.ResponseWriter, req *http.Request) {
	loginFormData := LoginFormData{
		IsLoggedIn: false,
	}

	//Build index header
	loginFormTmpl := template.Must(template.ParseFiles("templates/loginForm.tmpl"))

	var loginFormBuffer bytes.Buffer
	err := loginFormTmpl.Execute(&loginFormBuffer, loginFormData)
	if err != nil {
		log.Printf("failed executing login form template: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	//Build index nav

	//Build index main

	//Build index footer

	//Build overall index page
	indexData := IndexPageData{
		HeaderHTML: loginFormBuffer.String(),
	}
	indexTmpl := template.Must(template.ParseFiles("templates/index.tmpl"))

	err = indexTmpl.Execute(w, indexData)
	if err != nil {
		log.Printf("failed executing index template: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}
