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
	options    MicOptions
}

type MicOptions struct {
	Device string
	Volume string
}

func NewCarMic(options *MicOptions) (*CarMic, error) {
	// Create a audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		return nil, fmt.Errorf("error creating audio track: %w", err)
	}

	carMic := CarMic{
		AudioTrack: audioTrack,
		options: MicOptions{
			Volume: DefaultVolume,
			Device: DefaultDevice,
		},
	}

	if options != nil {
		carMic.options = *options
	}

	return &carMic, nil
}

func (c *CarMic) Start() {
	log.Println("Creating Pipeline")
	gst.CreateMicSendPipeline([]*webrtc.TrackLocalStaticSample{c.AudioTrack}, c.options.Volume).Start()
}
