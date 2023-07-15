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

	go StartDataListener(ctx, c.DataChannel, c.AudioTracks[0])

	StartGoGST(ctx, c.DataChannel)
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
	firstVideoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/h264"}, "video", "pion2")
	if err != nil {
		return fmt.Errorf("error creating first video track: %w", err)
	}
	c.VideoTracks = append(c.VideoTracks, firstVideoTrack)

	// Create a second video track
	secondVideoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/h264"}, "video", "pion3")
	if err != nil {
		return fmt.Errorf("error creating second video track: %w", err)
	}

	c.VideoTracks = append(c.VideoTracks, secondVideoTrack)
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
			err := track.WriteSample(media.Sample{Data: data, Duration: time.Duration(time.Nanosecond * 48000)})
			if err != nil {
				log.Printf("error writing sample to track: %s\n", err.Error())
			}
		}
	}
}

// func (c *CarCam) startStreaming() {
// 	fmt.Printf("%s starting streams...\n", c.Name)
// 	audioSrc := "audiotestsrc" //audiotestsrc
// 	c.AudioPipeline = gst.CreatePipeline("opus", []*webrtc.TrackLocalStaticSample{c.AudioTracks[0]}, audioSrc)
// 	c.AudioPipeline.Start()

// 	//libcamerasrc ! video/x-raw, width=640, height=480, framerate=30/1 ! videoconvert
// 	//autovideosrc ! video/x-raw, width=320, height=240 ! videoconvert ! queue

// 	//videoSrc := "libcamerasrc ! video/x-raw,format=YUY2,height=480,width=640,colorimetry=2:4:5:1,framerate=30/1 ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3,stream-forrmat=byte-stream,alignment=au,profiile=baseline,width=640,height=480,pixel-aspect-ratio=1/1,colorimetry=bt709,interlace-mode=progressive'" //webcam
// 	//videoSrc := "libcamerasrc ! video/x-raw,format=YUY2,height=480,width=640,colorimetry=2:4:5:1,framerate=30/1 ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3,stream-forrmat=byte-stream,alignment=au,profiile=baseline,width=640,height=480,pixel-aspect-ratio=1/1,colorimetry=bt709,interlace-mode=progressive'"
// 	//videoSrc := "libcamerasrc ! video/x-raw,format=YUY2,height=480,width=640,colorimetry=2:4:5:1,framerate=30/1 ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3,stream-format=byte-stream'"
// 	videoSrc := "libcamerasrc ! video/x-raw,format=YUY2,height=480,width=640,colorimetry=2:4:5:1,framerate=10/1 ! videoconvert ! v4l2h264enc ! video/x-h264, stream-format=byte-stream,level=3,alignment=au,profiile=baseline,width=640,height=480,pixel-aspect-ratio=1/1,colorimetry=bt709,interlace-mode=progressive"

// 	//videoSrc := "libcamerasrc"
// 	c.VideoPipeline = gst.CreatePipeline("custom", []*webrtc.TrackLocalStaticSample{c.VideoTracks[0], c.VideoTracks[1]}, videoSrc)
// 	c.VideoPipeline.Start()
// }

// func (c *CarCam) stopStreaming() {
// 	fmt.Printf("%s stopping streams...\n", c.Name)
// 	if c.AudioPipeline != nil {
// 		c.AudioPipeline.Stop()
// 	}

// 	if c.VideoPipeline != nil {
// 		c.VideoPipeline.Stop()
// 	}
// }
