package carcam

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type CarCam struct {
	VideoTrack   *webrtc.TrackLocalStaticSample
	videoChannel chan []byte
	options      CameraOptions
}

type CameraOptions struct {
	Name           string
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
}

func NewCarCam(options CameraOptions) (*CarCam, error) {
	// Create a video track
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if err != nil {
		return nil, fmt.Errorf("error creating first video track: %w", err)
	}

	return &CarCam{
		VideoTrack:   videoTrack,
		videoChannel: make(chan []byte, 5),
		options: CameraOptions{
			Name:           options.Name,
			Width:          options.Width,
			Height:         options.Height,
			Fps:            options.Fps,
			HorizontalFlip: options.HorizontalFlip,
			VerticalFlip:   options.VerticalFlip,
			DeNoise:        options.DeNoise,
			Rotation:       options.Rotation,
			Level:          "4.2",
			Profile:        options.Profile, //baseline, main or high
		},
	}, nil
}

func (c *CarCam) Start(ctx context.Context) error {
	go c.StartVideoDataListener(ctx)
	return c.StartStreaming(ctx)
}

func (c *CarCam) StartVideoDataListener(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Data Listener Done due to ctx")
			return
		case data, ok := <-c.videoChannel:
			if !ok {
				log.Println("Data channel closed, stopping")
				return
			}
			err := c.VideoTrack.WriteSample(media.Sample{Data: data, Duration: time.Millisecond * 17}) //TODO: Tie this to FPS
			if err != nil {
				log.Printf("error writing sample to track: %s\n", err.Error())
				return
			}
		}
	}
}
