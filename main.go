package main

import (
	"context"
	"log"
	"net/http"

	carcam "github.com/Speshl/goremotecontrol_web/internal/carcam"
	"github.com/Speshl/goremotecontrol_web/internal/server"
)

func main() {
	log.Println("Starting server...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Temp way to connect client to server before splitting client out to separate repo
	carCam := carcam.NewCarCam("Car-Alpha")
	socketServer := server.NewServer(carCam)
	socketServer.RegisterHTTPHandlers()
	socketServer.RegisterSocketIOHandlers()

	defer socketServer.Close()

	go func() {
		log.Println("Starting CarCam Client...")
		if err := carCam.ListenAndServe(ctx); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()

	go func() {
		log.Println("Start serving socketio...")
		if err := socketServer.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()

	log.Println("Start serving http...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
