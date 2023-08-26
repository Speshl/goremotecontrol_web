package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Speshl/goremotecontrol_web/internal/carcam"
	"github.com/Speshl/goremotecontrol_web/internal/carcommand"
	"github.com/Speshl/goremotecontrol_web/internal/carmic"
	"github.com/Speshl/goremotecontrol_web/internal/carspeaker"
	"github.com/Speshl/goremotecontrol_web/internal/gst"
	"github.com/Speshl/goremotecontrol_web/internal/server"
)

func (a *App) StartGStreamerPipelines() {
	go func() {
		log.Println("starting gstreamer main send recieve loops")
		gst.StartMainSendLoop() //Start gstreamer main send loop from main thread
		log.Println("starting gstreamer main recieve loops")
		gst.StartMainRecieveLoop() //Start gstreamer main recieve loop from main thread
		log.Println("warning: gstreamer main loops ended")
	}()
}

func (a *App) StartSpeaker() (*carspeaker.CarSpeaker, error) {
	carSpeaker, err := carspeaker.NewCarSpeaker(a.config.SpeakerConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating carspeaker: %w\n", err)
	}

	go func() {
		err = carSpeaker.Start(a.ctx)
		if err != nil {
			log.Printf("carspeaker error: %s\n", err.Error())
		}
		a.cancel() //stop anything else on this context because mic stopped
		a.done <- os.Kill
		log.Println("Stopping due to carspeaker stopping unexpectedly")
	}()

	return carSpeaker, nil
}

func (a *App) StartMic() (*carmic.CarMic, error) {
	carMic, err := carmic.NewCarMic(a.config.MicConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating carmic: %w\n", err)
	}

	carMic.Start()
	return carMic, nil
}

func (a *App) StartCam() (*carcam.CarCam, error) {
	carCam, err := carcam.NewCarCam(a.config.CamConfig)
	if err != nil {
		fmt.Errorf("error creating carcam: %w\n", err)
	}

	go func() {
		err = carCam.Start(a.ctx)
		if err != nil {
			log.Printf("carcam error: %s\n", err.Error())
		}
		a.cancel() //stop anything else on this context because camera stopped
		a.done <- os.Kill
		log.Println("Stopping due to carcam stopping unexpectedly")
	}()
	return carCam, nil
}

func (a *App) StartCommand() *carcommand.CarCommand {
	carCommand := carcommand.NewCarCommand(a.config.CommandConfig)
	go func() {
		err := carCommand.Start(a.ctx)
		if err != nil {
			log.Printf("carcommand error: %s\n", err.Error())
		}
		a.cancel() //stop anything else on this context because the gpio output stopped
		a.done <- os.Kill
		log.Println("Stopping due to carcommand stopping unexpectedly")
	}()

	return carCommand
}

func (a *App) StartSocketServer() *server.Server {
	socketServer := server.NewSocketServer(
		a.mic.AudioTrack,
		a.cam.VideoTrack,
		a.command.CommandChannel,
		a.speaker.SpeakerChannel,
		a.config.SpeakerConfig.Device,
		a.config.SpeakerConfig.Volume,
	)
	socketServer.RegisterHTTPHandlers()
	socketServer.RegisterSocketIOHandlers()

	go func() {
		log.Println("Start serving socketio...")
		if err := socketServer.Serve(); err != nil {
			log.Printf("socketio listen error: %s\n", err)
		}
		a.cancel() //stop anything else on this context because the socker server stopped
		a.done <- os.Kill
		log.Println("Stopping due to socker server stopping unexpectedly")
	}()

	return socketServer
}

func (a *App) StartHTTPServer() {
	go func() {
		log.Println("Start serving http...")
		addr := fmt.Sprintf(":%s", a.config.ServerConfig.Port)
		err := http.ListenAndServe(addr, nil)
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
		a.cancel() //stop anything else on this context because the http server stopped
		a.done <- os.Kill
		log.Println("Stopping due to http server stopping unexpectedly")
	}()
}
