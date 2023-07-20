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
	http.HandleFunc("/video_test", s.streamHandler)

	http.Handle("/socket.io/", s.socketio)
}

func (s *Server) driveHandler(w http.ResponseWriter, req *http.Request) {
	// user := req.PostFormValue("user")
	// password := req.PostFormValue("password")

	// if user != "user" || password != "password" {
	// 	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	// 	return
	// }

	template := template.Must(template.ParseFiles("static/html/drive.html"))
	template.Execute(w, nil) //Can pass map[string]any here and use go templates to dynamically build the html page
}

func (s *Server) loginHandler(w http.ResponseWriter, req *http.Request) {
	template := template.Must(template.ParseFiles("static/html/login.html"))
	template.Execute(w, nil)
}

func (s *Server) gptHandler(w http.ResponseWriter, req *http.Request) {
	template := template.Must(template.ParseFiles("static/html/gpt.html"))
	template.Execute(w, nil)
}

func (s *Server) streamHandler(w http.ResponseWriter, req *http.Request) {
	template := template.Must(template.ParseFiles("static/html/stream.html"))
	template.Execute(w, nil)
}

func (s *Server) videoTestHandler(w http.ResponseWriter, req *http.Request) {
	template := template.Must(template.ParseFiles("static/html/video_test.html"))
	template.Execute(w, nil)
}
