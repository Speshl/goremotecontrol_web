package carmic

import (
	"fmt"
	"log"

	"github.com/Speshl/goremotecontrol_web/internal/gst"
	"github.com/pion/webrtc/v3"
)

const DefaultDevice = "0"
const DefaultVolume = "5.0"

type CarMic struct {
	AudioTrack *webrtc.TrackLocalStaticSample
	config     MicConfig
}

type MicConfig struct {
	Device string
	Volume string
}

func NewCarMic(cfg MicConfig) (*CarMic, error) {
	// Create a audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		return nil, fmt.Errorf("error creating audio track: %w", err)
	}

	carMic := CarMic{
		AudioTrack: audioTrack,
		config:     cfg,
	}

	return &carMic, nil
}

func (c *CarMic) Start() {
	log.Println("Creating Mic Pipeline")
	gst.CreateMicSendPipeline([]*webrtc.TrackLocalStaticSample{c.AudioTrack}, c.config.Device, c.config.Volume).Start()
}
