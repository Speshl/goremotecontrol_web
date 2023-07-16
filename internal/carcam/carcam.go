package carcam

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Speshl/goremotecontrol_web/internal/carcam/gst"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type CarCam struct {
	Name        string
	AudioTracks []*webrtc.TrackLocalStaticSample
	VideoTracks []*webrtc.TrackLocalStaticSample
	DataChannel chan []byte

	AudioPipeline *gst.Pipeline
	VideoPipeline *gst.Pipeline
}

func NewCarCam(name string) *CarCam {
	return &CarCam{
		Name:        name,
		AudioTracks: make([]*webrtc.TrackLocalStaticSample, 0),
		VideoTracks: make([]*webrtc.TrackLocalStaticSample, 0),
		DataChannel: make(chan []byte, 5),
	}
}

func (c *CarCam) ListenAndServe(ctx context.Context) error {
	err := c.CreateTracks()
	if err != nil {
		return err
	}

	go StartDataListener(ctx, c.DataChannel, c.VideoTracks[0])
	StartStreaming(ctx, c.DataChannel)
	return nil
}

func (c *CarCam) CreateTracks() error {
	log.Printf("%s started creating tracks...", c.Name)
	defer log.Printf("%s finished creating tracks", c.Name)

	// Create a audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		return fmt.Errorf("error creating audio track: %w", err)
	}
	c.AudioTracks = append(c.AudioTracks, audioTrack)

	// Create a video track
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if err != nil {
		return fmt.Errorf("error creating first video track: %w", err)
	}
	c.VideoTracks = append(c.VideoTracks, videoTrack)
	return nil
}

func StartDataListener(ctx context.Context, dataChannel chan []byte, track *webrtc.TrackLocalStaticSample) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Data Listener Done due to ctx")
			return
		case data, ok := <-dataChannel:
			if !ok {
				log.Println("Data channel closed, stopping")
			}
			log.Println("writing data to track")
			err := track.WriteSample(media.Sample{Data: data, Duration: time.Millisecond * 17})
			if err != nil {
				log.Printf("error writing sample to track: %s\n", err.Error())
			}
		}
	}
}
