package main

import (
	"context"
	"log"
	"net/http"

	carcam "github.com/Speshl/goremotecontrol_web/internal/carcam"
	"github.com/Speshl/goremotecontrol_web/internal/carcommand"
	"github.com/Speshl/goremotecontrol_web/internal/server"
)

const carName = "Car-Alpha"

const width = "1280"
const height = "720"
const fps = "60"
const refreshRate = 60 //command refresh rate

func main() {
	log.Println("Starting server...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Temp way to connect client to server before splitting client out to separate repo
	carCam, err := carcam.NewCarCam(carName, width, height, fps)
	if err != nil {
		log.Fatalf("NewCarCam error: %s\n", err)
	}

	err = carCam.Start(ctx)
	if err != nil {
		log.Fatalf("carcam error: %s\n", err.Error())
	}

	carCommand := carcommand.NewCarCommand(carName, refreshRate)
	carCommand.Start(ctx)

	socketServer := server.NewServer(carCam)
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
