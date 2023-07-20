package server

import (
	"html/template"
	"net/http"
)

func (s *Server) RegisterHTTPHandlers() {
	http.Handle("/", http.FileServer(http.Dir("public/")))
	http.Handle("/socket.io/", s.socketio)

}

func (s *Server) driveHandler(w http.ResponseWriter, req *http.Request) {
	// user := req.PostFormValue("user")
	// password := req.PostFormValue("password")

	// if user != "user" || password != "password" {
	// 	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	// 	return
	// }

	template := template.Must(template.ParseFiles("public/html/drive.html"))
	template.Execute(w, nil) //Can pass map[string]any here and use go templates to dynamically build the html page
}
