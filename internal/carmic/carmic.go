package carmic

import (
	"fmt"
	"log"

	"github.com/Speshl/goremotecontrol_web/internal/gst"
	"github.com/pion/webrtc/v3"
)

type CarMic struct {
	AudioTrack *webrtc.TrackLocalStaticSample
	options    MicOptions
}

type MicOptions struct {
	Name string
}

func NewCarMic(options MicOptions) (*CarMic, error) {
	// Create a audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		return nil, fmt.Errorf("error creating audio track: %w", err)
	}

	//gst.InitSendLoop()

	return &CarMic{
		AudioTrack: audioTrack,
		options:    options,
	}, nil
}

func (c *CarMic) Start() {
	log.Println("Creating Pipeline")
	gst.CreateMicSendPipeline([]*webrtc.TrackLocalStaticSample{c.AudioTrack}).Start()
}
