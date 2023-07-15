package carcam

import (
	"fmt"
	"log"

	"github.com/Speshl/goremotecontrol_web/internal/server/gst"
	"github.com/pion/webrtc/v3"
)

type CarCam struct {
	Name        string
	AudioTracks []*webrtc.TrackLocalStaticSample
	VideoTracks []*webrtc.TrackLocalStaticSample

	AudioPipeline *gst.Pipeline
	VideoPipeline *gst.Pipeline
}

func NewCarCam(name string) *CarCam {
	return &CarCam{
		Name:        name,
		AudioTracks: make([]*webrtc.TrackLocalStaticSample, 0),
		VideoTracks: make([]*webrtc.TrackLocalStaticSample, 0),
	}
}

func (c *CarCam) ListenAndServe() error {
	err := c.CreateTracks()
	if err != nil {
		return err
	}

	c.startStreaming()
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
	firstVideoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion2")
	if err != nil {
		return fmt.Errorf("error creating first video track: %w", err)
	}
	c.VideoTracks = append(c.VideoTracks, firstVideoTrack)

	// Create a second video track
	secondVideoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion3")
	if err != nil {
		return fmt.Errorf("error creating second video track: %w", err)
	}

	c.VideoTracks = append(c.VideoTracks, secondVideoTrack)
	return nil
}

func (c *CarCam) startStreaming() {
	fmt.Printf("%s starting streams...\n", c.Name)
	audioSrc := "audiotestsrc" //audiotestsrc
	c.AudioPipeline = gst.CreatePipeline("opus", []*webrtc.TrackLocalStaticSample{c.AudioTracks[0]}, audioSrc)
	c.AudioPipeline.Start()

	//libcamerasrc ! video/x-raw, width=640, height=480, framerate=30/1 ! videoconvert
	//autovideosrc ! video/x-raw, width=320, height=240 ! videoconvert ! queue

	videoSrc := "libcamerasrc ! video/x-raw, width=640, height=480, framerate=30/1 ! videoconvert ! queue" //webcam
	//videoSrc := "videotestsrc"
	c.VideoPipeline = gst.CreatePipeline("vp8", []*webrtc.TrackLocalStaticSample{c.VideoTracks[0], c.VideoTracks[1]}, videoSrc)
	c.VideoPipeline.Start()
}

func (c *CarCam) stopStreaming() {
	fmt.Printf("%s stopping streams...\n", c.Name)
	if c.AudioPipeline != nil {
		c.AudioPipeline.Stop()
	}

	if c.VideoPipeline != nil {
		c.VideoPipeline.Stop()
	}
}

//gstreamer tests
/*
image/jpeg, width=640, height=480

gst-launch-1.0 libcamerasrc ! jpegdec ! videoconvert ! queue ! vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1 ! testsink

gst-launch-1.0 libcamerasrc ! jpegdec ! videoconvert ! queue ! filesink location=gstreamer_capture

gst-launch-1.0 libcamerasrc ! jpegdec ! videoconvert ! filesink location=gstreamer_capture

gst-launch-1.0 -v filesrc location=mjpeg.avi ! avidemux !  queue ! jpegdec ! videoconvert ! videoscale ! autovideosink





video/x-raw, format=YUY2, width=1280, height=960

gst-launch-1.0 libcamerasrc ! video/x-raw, width=640, height=480, framerate=30/1 ! videoconvert ! queue ! vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1 ! testsink

*/
