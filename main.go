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

	"github.com/Speshl/goremotecontrol_web/internal/carcam"
	"github.com/Speshl/goremotecontrol_web/internal/carcommand"
	"github.com/Speshl/goremotecontrol_web/internal/carmic"
	"github.com/Speshl/goremotecontrol_web/internal/carspeaker"
	"github.com/Speshl/goremotecontrol_web/internal/gst"
	"github.com/Speshl/goremotecontrol_web/internal/server"
)

func main() {
	log.Println("starting server...")
	defer log.Println("server stopped")
	ctx, cancel := context.WithCancel(context.Background())

	carConfig := GetConfig(ctx)

	//Start client audio pipeline listener
	go func() {
		log.Println("starting gstreamer main loops")
		gst.StartMainRecieveLoop() //Start gstreamer main loop from main thread
		log.Println("warning: gstreamer main loop ended")
	}()

	time.Sleep(2 * time.Second)

	carSpeaker, err := carspeaker.NewCarSpeaker(carConfig.speakerConfig)
	if err != nil {
		log.Printf("NewCarCam error: %s\n", err)
	}

	// go func() {
	// 	err = carSpeaker.Start(ctx)
	// 	if err != nil {
	// 		log.Printf("carspeaker error: %s\n", err.Error())
	// 	}
	// 	cancel() //stop anything else on this context because mic stopped
	// 	log.Println("Stopping due to carspeaker stopping unexpectedly")
	// }()

	// go func() {
	// 	err = carSpeaker.Play(ctx, "startup")
	// 	if err != nil {
	// 		log.Printf("caraudio error: %s\n", err.Error())
	// 	}
	// }()

	carMic, err := carmic.NewCarMic(carConfig.micConfig)
	if err != nil {
		log.Printf("NewCarMic error: %s\n", err)
		cancel() //stop anything else on this context because mic stopped
	}

	carMic.Start()

	// go func() {
	// 	err = carMic.Start(ctx)
	// 	if err != nil {
	// 		log.Printf("carmic error: %s\n", err.Error())
	// 	}
	// 	cancel() //stop anything else on this context because mic stopped
	// 	log.Println("Stopping due to carmic stopping unexpectedly")
	// }()

	//Temp way to connect client to server before splitting client out to separate repo
	carCam, err := carcam.NewCarCam(carConfig.camConfig)
	if err != nil {
		log.Printf("NewCarCam error: %s\n", err)
		cancel() //stop anything else on this context because camera stopped
	}

	go func() {
		err = carCam.Start(ctx)
		if err != nil {
			log.Printf("carcam error: %s\n", err.Error())
		}
		cancel() //stop anything else on this context because camera stopped
		log.Println("Stopping due to carcam stopping unexpectedly")
	}()

	//give time for camera to start before commands start
	time.Sleep(2 * time.Second)

	carCommand := carcommand.NewCarCommand(carConfig.commandConfig)
	go func() {
		err := carCommand.Start(ctx)
		if err != nil {
			log.Printf("carcommand error: %s\n", err.Error())
		}
		cancel() //stop anything else on this context because the gpio output stopped
		log.Println("Stopping due to carcommand stopping unexpectedly")
	}()

	socketServer := server.NewServer(carMic.AudioTrack, carCam.VideoTrack, carCommand.CommandChannel, carSpeaker.SpeakerChannel, carMic)
	socketServer.RegisterHTTPHandlers()
	socketServer.RegisterSocketIOHandlers()

	defer socketServer.Close()

	go func() {
		log.Println("Start serving socketio...")
		if err := socketServer.Serve(); err != nil {
			log.Printf("socketio listen error: %s\n", err)
		}
		cancel() //stop anything else on this context because the socker server stopped
		log.Println("Stopping due to socker server stopping unexpectedly")
	}()

	go func() {
		log.Println("Start serving http...")
		err = http.ListenAndServe(":8181", nil)
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
		cancel() //stop anything else on this context because the http server stopped
		log.Println("Stopping due to http server stopping unexpectedly")
	}()

	//Handle shutdown signals
	signal.Ignore(os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan os.Signal, 1)
	defer close(done)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case msg := <-done:
			log.Printf("Shutting down server... %s\n", msg.String())
			cancel()
			//give some time for everything to close
			time.Sleep(5 * time.Second)
			return
		}
	}
}
