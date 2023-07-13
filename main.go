package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/Speshl/goremotecontrol_web/server"
)

func main() {
	driveHandler := func(w http.ResponseWriter, req *http.Request) {
		user := req.PostFormValue("user")
		password := req.PostFormValue("password")

		if user != "user" || password != "password" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		template := template.Must(template.ParseFiles("html/drive.html"))
		template.Execute(w, nil) //Can pass map[string]any here and use go templates to dynamically build the html page
	}

	loginHandler := func(w http.ResponseWriter, req *http.Request) {
		template := template.Must(template.ParseFiles("html/login.html"))
		template.Execute(w, nil)
	}

	gptHandler := func(w http.ResponseWriter, req *http.Request) {
		template := template.Must(template.ParseFiles("html/gpt.html"))
		template.Execute(w, nil)
	}

	audioHandler := func(w http.ResponseWriter, req *http.Request) {
		template := template.Must(template.ParseFiles("html/audio.html"))
		template.Execute(w, nil)
	}

	socketServer := server.NewSocketServer()
	socketServer.RegisterHandlers()

	go func() {
		if err := socketServer.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
	defer socketServer.Close()

	http.HandleFunc("/drive", driveHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/gpt", gptHandler)
	http.HandleFunc("/audio", audioHandler)

	http.Handle("/socket.io/", socketServer.GetHandler())

	err := server.TempRecordCam()
	if err != nil {
		log.Fatalf("Failed temp record cam: %s\n", err.Error())
	}

	log.Println("Start serving...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
