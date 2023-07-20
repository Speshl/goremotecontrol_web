package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	carcam "github.com/Speshl/goremotecontrol_web/internal/carcam"
	"github.com/Speshl/goremotecontrol_web/internal/carcommand"
	"github.com/Speshl/goremotecontrol_web/internal/server"
)

const carName = "Car-Alpha"

const width = "640"
const height = "480"
const fps = "60"
const refreshRate = 60 //command refresh rate

func main() {
	log.Println("Starting server...")
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		log.Println("Stopping server...")
		cancel()
		time.Sleep(5 * time.Second)
	}()

	//Temp way to connect client to server before splitting client out to separate repo
	carCam, err := carcam.NewCarCam(carName, width, height, fps)
	if err != nil {
		log.Fatalf("NewCarCam error: %s\n", err)
	}

	go func() {
		err = carCam.Start(ctx)
		if err != nil {
			log.Fatalf("carcam error: %s\n", err.Error())
		}
		cancel() //stop anything else on this context because camera stopped
		log.Println("Stopping due to carcam stopping unexpectedly")
	}()

	carCommand := carcommand.NewCarCommand(carName, refreshRate)
	go func() {
		err := carCommand.Start(ctx)
		if err != nil {
			log.Fatalf("carcommand error: %s\n", err.Error())
		}
		cancel() //stop anything else on this context because the gpio output stopped
		log.Println("Stopping due to carcommand stopping unexpectedly")
	}()

	socketServer := server.NewServer(carCam, carCommand)
	socketServer.RegisterHTTPHandlers()
	socketServer.RegisterSocketIOHandlers()

	defer socketServer.Close()

	go func() {
		log.Println("Start serving socketio...")
		if err := socketServer.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()

	go func() {
		log.Println("Start serving http...")
		err = http.ListenAndServe(":8181", nil)
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	signal.Ignore(os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan os.Signal, 1)
	defer close(done)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer log.Println("Starting close process")
	for {
		select {
		case msg := <-done:
			log.Printf("Shutting down server... %s\n", msg.String())
			cancel()
			return
		}
	}
}
