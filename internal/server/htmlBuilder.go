package server

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
)

type IndexBuildOptions struct {
	includeShell bool
	authorized   bool
	userName     string
	userRank     string
}

type PageShellData struct {
	Body template.HTML
}

type IndexBodyData struct {
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

func (s *Server) buildIndex(w http.ResponseWriter, options IndexBuildOptions) {
	loginFormData := LoginFormData{
		IsLoggedIn: options.authorized,
		Username:   options.userName,
		Rank:       options.userRank,
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

	//Build overall index body
	indexBodyData := IndexBodyData{
		HeaderHTML: template.HTML(loginFormBuffer.String()),
	}
	indexBodyTmpl := template.Must(template.ParseFiles("templates/index.tmpl"))

	if !options.includeShell {
		err = indexBodyTmpl.Execute(w, indexBodyData)
		if err != nil {
			log.Printf("failed executing index body: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	//place into shell

	var indexBodyBuffer bytes.Buffer
	err = indexBodyTmpl.Execute(&indexBodyBuffer, indexBodyData)
	if err != nil {
		log.Printf("failed executing index body: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	//Apply body to page shell
	indexShellData := PageShellData{
		Body: template.HTML(indexBodyBuffer.String()),
	}
	indexTmpl := template.Must(template.ParseFiles("templates/index.tmpl"))

	var indexBuffer bytes.Buffer
	err = indexTmpl.Execute(&indexBuffer, indexShellData)
	if err != nil {
		log.Printf("failed executing index shell: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}
