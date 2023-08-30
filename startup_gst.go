package main

import (
	"log"

	"github.com/Speshl/goremotecontrol_web/internal/gst"
)

// Pulled out because this file shows errors due to gst and cgo
func (a *App) StartGStreamerPipelines() {
	go func() {
		log.Println("starting gstreamer main send recieve loops")
		gst.StartMainSendLoop() //Start gstreamer main send loop from main thread
		log.Println("starting gstreamer main recieve loops")
		gst.StartMainRecieveLoop() //Start gstreamer main recieve loop from main thread
		log.Println("warning: gstreamer main loops ended")
	}()
}
