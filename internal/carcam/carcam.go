package carcam

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

const DefaultLevel = "4.2"
const DefaultFPS = 30

type CarCam struct {
	VideoTrack   *webrtc.TrackLocalStaticSample
	videoChannel chan []byte
	config       CamConfig
}

type CamConfig struct {
	Width          string
	Height         string
	Fps            string
	DisableVideo   bool
	HorizontalFlip bool
	VerticalFlip   bool
	DeNoise        bool
	Rotation       int
	Level          string
	Profile        string
	Mode           string
}

func NewCarCam(cfg CamConfig) (*CarCam, error) {
	// Create a video track
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if err != nil {
		return nil, fmt.Errorf("error creating first video track: %w", err)
	}

	carCam := CarCam{
		VideoTrack:   videoTrack,
		videoChannel: make(chan []byte, 5),
		config:       cfg,
	}
	carCam.config.Level = DefaultLevel
	return &carCam, nil
}

func (c *CarCam) Start(ctx context.Context) error {
	go c.StartVideoDataListener(ctx)
	return c.StartStreaming(ctx)
}

func (c *CarCam) StartVideoDataListener(ctx context.Context) {
	fps, err := strconv.ParseInt(c.config.Fps, 10, 32)
	if err != nil {
		fps = DefaultFPS
	}

	duration := int(1000 / fps)
	for {
		select {
		case <-ctx.Done():
			log.Println("video data listener done due to ctx")
			return
		case data, ok := <-c.videoChannel:
			if !ok {
				log.Println("video data channel closed, stopping")
				return
			}

			err := c.VideoTrack.WriteSample(media.Sample{Data: data, Duration: time.Millisecond * time.Duration(duration)})
			if err != nil {
				log.Printf("error writing sample to track: %s\n", err.Error())
				return
			}
		}
	}
}
