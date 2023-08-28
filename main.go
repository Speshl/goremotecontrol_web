package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Speshl/goremotecontrol_web/internal/carcam"
	"github.com/Speshl/goremotecontrol_web/internal/carcommand"
	"github.com/Speshl/goremotecontrol_web/internal/carmic"
	"github.com/Speshl/goremotecontrol_web/internal/carspeaker"
	"github.com/Speshl/goremotecontrol_web/internal/config"
	"github.com/Speshl/goremotecontrol_web/internal/server"
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc
	done   chan os.Signal
	config config.CarConfig

	speaker      *carspeaker.CarSpeaker
	mic          *carmic.CarMic
	cam          *carcam.CarCam
	command      *carcommand.CarCommand
	socketServer *server.Server
}

func main() {
	log.Println("starting server...")
	defer log.Println("server stopped")

	app := App{
		done: make(chan os.Signal, 1),
	}
	defer close(app.done)

	app.ctx, app.cancel = context.WithCancel(context.Background())
	app.config = config.GetConfig(app.ctx)

	//Start audio recieve pipeline listener
	app.StartGStreamerPipelines()

	carspeaker, err := app.StartSpeaker()
	if err != nil {
		app.cancel()
		app.done <- os.Kill
		log.Fatalf("failed starting speaker - %w", err)
	}
	app.speaker = carspeaker

	//Play startup sound
	// go func() {
	// 	err = app.speaker.Play(app.ctx, "startup")
	// 	if err != nil {
	// 		log.Printf("caraudio error: %s\n", err.Error())
	// 	}
	// }()

	carmic, err := app.StartMic()
	if err != nil {
		app.cancel()
		app.done <- os.Kill
		log.Fatalf("failed starting mic - %w", err)
	}
	app.mic = carmic

	carCam, err := app.StartCam()
	if err != nil {
		app.cancel()
		app.done <- os.Kill
		log.Fatalf("failed starting mic - %w", err)
	}
	app.cam = carCam

	//give time for camera to start before commands start
	time.Sleep(2 * time.Second)

	app.command = app.StartCommand()

	app.socketServer = app.StartSocketServer()
	defer app.socketServer.Close()

	app.StartHTTPServer()

	//Handle shutdown signals
	signal.Ignore(os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(app.done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	for { //Block here until the world ends OR we get a kill signal
		select {
		case msg := <-app.done:
			log.Printf("Shutting down server... %s\n", msg.String())
			app.cancel()
			//give some time for everything to close
			time.Sleep(5 * time.Second)
			return
		}
	}
}
