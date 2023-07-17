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
	Name         string
	AudioTrack   *webrtc.TrackLocalStaticSample
	VideoTrack   *webrtc.TrackLocalStaticSample
	videoChannel chan []byte
	options      CameraOptions
}

type CameraOptions struct {
	width          string
	height         string
	fps            string
	horizontalFlip bool
	verticalFlip   bool
	deNoise        bool
	rotation       int
	level          string
	profile        string
}

func NewCarCam(name string, width string, height string, fps string) (*CarCam, error) {
	// Create a audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		return nil, fmt.Errorf("error creating audio track: %w", err)
	}

	// Create a video track
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if err != nil {
		return nil, fmt.Errorf("error creating first video track: %w", err)
	}

	return &CarCam{
		Name:         name,
		AudioTrack:   audioTrack,
		VideoTrack:   videoTrack,
		videoChannel: make(chan []byte, 5),
		options: CameraOptions{
			width:          width,
			height:         height,
			fps:            fps,
			horizontalFlip: false,
			verticalFlip:   false,
			deNoise:        true,
			rotation:       0,
			level:          "4.2",
			profile:        "baseline", //baseline, main or high
		},
	}, nil
}

func (c *CarCam) Start(ctx context.Context) error {
	err := c.CreateTracks()
	if err != nil {
		return err
	}

	go c.StartVideoDataListener(ctx)
	c.StartStreaming(ctx)
	return nil
}

func (c *CarCam) CreateTracks() error {
	log.Printf("%s started creating tracks...", c.Name)
	defer log.Printf("%s finished creating tracks", c.Name)

	// Create a audio track
	var err error
	c.AudioTrack, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		return fmt.Errorf("error creating audio track: %w", err)
	}

	// Create a video track
	c.VideoTrack, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if err != nil {
		return fmt.Errorf("error creating first video track: %w", err)
	}
	return nil
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
			err := c.VideoTrack.WriteSample(media.Sample{Data: data, Duration: time.Millisecond * 17})
			if err != nil {
				log.Printf("error writing sample to track: %s\n", err.Error())
				return
			}
		}
	}
}
