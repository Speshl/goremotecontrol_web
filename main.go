package main

import (
	"log"
	"net/http"

	"github.com/Speshl/goremotecontrol_web/internal/server"
)

func main() {
	socketServer := server.NewServer()
	socketServer.RegisterHTTPHandlers()
	socketServer.RegisterSocketIOHandlers()

	defer socketServer.Close()

	go func() {
		log.Println("Start serving socketio...")
		if err := socketServer.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()

	log.Println("Start serving http...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
