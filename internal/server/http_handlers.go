package server

import (
	"html/template"
	"net/http"
)

func (s *Server) RegisterHTTPHandlers() {
	http.HandleFunc("/drive", s.driveHandler)
	http.HandleFunc("/login", s.loginHandler)
	http.HandleFunc("/gpt", s.gptHandler)
	http.HandleFunc("/stream", s.streamHandler)

	http.Handle("/socket.io/", s.socketio)
}

func (s *Server) driveHandler(w http.ResponseWriter, req *http.Request) {
	user := req.PostFormValue("user")
	password := req.PostFormValue("password")

	if user != "user" || password != "password" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	template := template.Must(template.ParseFiles("html/drive.html"))
	template.Execute(w, nil) //Can pass map[string]any here and use go templates to dynamically build the html page
}

func (s *Server) loginHandler(w http.ResponseWriter, req *http.Request) {
	template := template.Must(template.ParseFiles("html/login.html"))
	template.Execute(w, nil)
}

func (s *Server) gptHandler(w http.ResponseWriter, req *http.Request) {
	template := template.Must(template.ParseFiles("html/gpt.html"))
	template.Execute(w, nil)
}

func (s *Server) streamHandler(w http.ResponseWriter, req *http.Request) {
	template := template.Must(template.ParseFiles("html/stream.html"))
	template.Execute(w, nil)
}
