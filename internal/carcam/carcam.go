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

func (c *CarCam) startStreaming() {
	fmt.Printf("%s starting streams...\n", c.Name)
	audioSrc := "audiotestsrc" //audiotestsrc
	c.AudioPipeline = gst.CreatePipeline("opus", []*webrtc.TrackLocalStaticSample{c.AudioTracks[0]}, audioSrc)
	c.AudioPipeline.Start()

	//libcamerasrc ! video/x-raw, width=640, height=480, framerate=30/1 ! videoconvert
	//autovideosrc ! video/x-raw, width=320, height=240 ! videoconvert ! queue

	//videoSrc := "libcamerasrc ! video/x-raw,format=YUY2,height=480,width=640,colorimetry=2:4:5:1,framerate=30/1 ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3,stream-forrmat=byte-stream,alignment=au,profiile=baseline,width=640,height=480,pixel-aspect-ratio=1/1,colorimetry=bt709,interlace-mode=progressive'" //webcam
	//videoSrc := "libcamerasrc ! video/x-raw,format=YUY2,height=480,width=640,colorimetry=2:4:5:1,framerate=30/1 ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3,stream-forrmat=byte-stream,alignment=au,profiile=baseline,width=640,height=480,pixel-aspect-ratio=1/1,colorimetry=bt709,interlace-mode=progressive'"
	videoSrc := "libcamerasrc ! video/x-raw,format=YUY2,height=480,width=640,colorimetry=2:4:5:1,framerate=30/1 ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3,stream-format=byte-stream'"

	//videoSrc := "libcamerasrc"
	c.VideoPipeline = gst.CreatePipeline("custom", []*webrtc.TrackLocalStaticSample{c.VideoTracks[0], c.VideoTracks[1]}, videoSrc)
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

gst-launch-1.0 libcamerasrc ! video/x-raw, width=640, height=480, framerate=30/1 ! videoconvert ! queue ! vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1 ! filesink location=gstreamer_capture

gst-launch-1.0 libcamerasrc ! video/x-raw, width=640, height=480, framerate=30/1 ! filesink location=gstreamer_capture

gst-launch-1.0 libcamerasrc ! filesink location=gstreamer_capture

gst-launch-1.0 libcamerasrc ! videoconvert ! x264enc ! flvmux  ! filesink location=gstreamer_capture.flv -e



'appsrc ! "video/x-raw,framerate=25/1,format=BGR,width=640,height=480" ! ' \
           'queue ! v4l2h264enc ! "video/x-h264,level=(string)4" ! h264parse ! ' \
           'rtph264pay ! gdppay ! tcpserversink host=0.0.0.0 port=7000 '


v4l2h264enc is hardware decoder
x264enc is software decoder


gst-launch-1.0 -vvvv libcamerasrc ! video/x-raw, format=YUY2, width=1280, height=720, framerate=30/1, colorimetry=2:4:5:1 ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3' ! filesink location=gstreamer_capture.h264

gst-launch-1.0 -vvvv videotestsrc ! video/x-raw, format=YUY2, width=1280, height=720, framerate=30/1 ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3' ! filesink location=gstreamer_capture.h264


gst-launch-1.0 -vvvv libcamerasrc ! video/x-raw, format=YUY2, width=320, height=240, framerate=30/1 ! v4l2h264enc ! 'video/x-h264,level=(string)4' ! filesink location=gstreamer_capture.h264

gst-launch-1.0 -vvvv libcamerasrc ! video/x-raw, format=YUY2, width=320, height=240, framerate=30/1 ! videoconvert ! x264enc ! 'video/x-h264,level=(string)4' ! filesink location=gstreamer_capture.h264

export GST_DEBUG=6 gst-launch-1.0 libcamerasrc ! capsfilter caps="video/x-raw,width=320,height=240,format=YUY2,framerate=1/1,interlace-mode=(string)progressive,colorimetry=bt709" ! v4l2h264enc extra-controls="controls,repeat_sequence_header=1,h264_profile=1,h264_level=11,video_bitrate=5000000,h264_i_frame_period=30,h264_minimum_qp_value=10" ! "video/x-h264,level=(string)4" ! testsink


#Working recording
gst-launch-1.0  -vvvv libcamerasrc ! videoconvert ! x264enc ! flvmux  ! filesink location=gstreamer_capture.flv -e

gst-launch-1.0 -vvvv libcamerasrc ! video/x-raw, format=YUY2, width=320, height=240, framerate=30/1 ! videoconvert ! x264enc ! 'video/x-h264,level=(string)3' ! filesink location=gstreamer_capture.h264

gst-launch-1.0 -vvvv videotestsrc ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3' ! filesink location=gstreamer_capture.h264





# Trying ONNN camera


# Trying New Logitech
gst-launch-1.0 -vvvv libcamerasrc ! video/x-raw, format=YUY2, height=480, width=640 ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3' ! filesink location=gstreamer_capture.h264


gst-launch-1.0 -vvvv libcamerasrc ! video/x-raw, format=YUY2, height=480, width=640 ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3' ! filesink location=gstreamer_capture.h264

gst-launch-1.0 -vvvv libcamerasrc ! video/x-raw, format=YUY2, height=1920, width=1080, colorimetry=2:4:5:1, framerate=30/1 ! v4l2h264enc ! 'video/x-h264,level=(string)3' ! filesink location=gstreamer_capture.h264
 caps = video/x-raw, format=(string)YUY2, width=(int)1920, height=(int)1080, colorimetry=(string)2:4:5:1, framerate=(fraction)30/1










# Video Test Src
gst-launch-1.0 -vvvv videotestsrc ! video/x-raw, format=YUY2, height=480, width=640 ! v4l2h264enc ! 'video/x-h264,level=(string)3' ! filesink location=gstreamer_capture.h264

gst-launch-1.0 -vvvv videotestsrc ! video/x-raw,format=I420 ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! video/x-h264,stream-format=byte-stream ! testsink



gst-launch-1.0 libcamerasrc ! video/x-raw,format=YUY2,height=480,width=640,colorimetry=2:4:5:1,framerate=30/1 ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3,stream-forrmat=byte-stream,alignment=au,profiile=baseline,width=640,height=480,pixel-aspect-ratio=1/1,colorimetry=bt709,interlace-mode=progressive' ! testsink



******
gst-launch-1.0 -vvvv libcamerasrc ! video/x-raw,format=YUY2,height=480,width=640,colorimetry=2:4:5:1,framerate=30/1 ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3,stream-forrmat=byte-stream,alignment=au,profiile=baseline,width=640,height=480,pixel-aspect-ratio=1/1,colorimetry=bt709,interlace-mode=progressive' ! filesink location=gstreamer_capture.h264



gst-launch-1.0 -vvvv libcamerasrc ! video/x-raw, format=YUY2, height=480, width=640 ! videoconvert ! v4l2h264enc ! 'video/x-h264,level=(string)3' ! filesink location=gstreamer_capture.h264
Setting pipeline to PAUSED ...
[0:19:05.777974630] [1173]  INFO Camera camera_manager.cpp:299 libcamera v0.0.4+22-923f5d70
Pipeline is live and does not need PREROLL ...
Pipeline is PREROLLED ...
Setting pipeline to PLAYING ...
New clock: GstSystemClock
[0:19:05.804374695] [1175]  INFO Camera camera.cpp:1028 configuring streams: (0) 640x480-YUYV
/GstPipeline:pipeline0/GstLibcameraSrc:libcamerasrc0.GstLibcameraPad:src: caps = video/x-raw, format=(string)YUY2, width=(int)640, height=(int)480, colorimetry=(string)2:4:5:1, framerate=(fraction)30/1
/GstPipeline:pipeline0/GstCapsFilter:capsfilter0.GstPad:src: caps = video/x-raw, format=(string)YUY2, width=(int)640, height=(int)480, colorimetry=(string)2:4:5:1, framerate=(fraction)30/1
/GstPipeline:pipeline0/GstVideoConvert:videoconvert0.GstPad:src: caps = video/x-raw, width=(int)640, height=(int)480, framerate=(fraction)30/1, format=(string)YUY2, interlace-mode=(string)progressive, colorimetry=(string)bt709
/GstPipeline:pipeline0/v4l2h264enc:v4l2h264enc0.GstPad:src: caps = video/x-h264, stream-format=(string)byte-stream, alignment=(string)au, level=(string)3, profile=(string)baseline, width=(int)640, height=(int)480, pixel-aspect-ratio=(fraction)1/1, framerate=(fraction)30/1, interlace-mode=(string)progressive, colorimetry=(string)bt709
/GstPipeline:pipeline0/GstCapsFilter:capsfilter1.GstPad:src: caps = video/x-h264, stream-format=(string)byte-stream, alignment=(string)au, level=(string)3, profile=(string)baseline, width=(int)640, height=(int)480, pixel-aspect-ratio=(fraction)1/1, framerate=(fraction)30/1, interlace-mode=(string)progressive, colorimetry=(string)bt709
/GstPipeline:pipeline0/GstFileSink:filesink0.GstPad:sink: caps = video/x-h264, stream-format=(string)byte-stream, alignment=(string)au, level=(string)3, profile=(string)baseline, width=(int)640, height=(int)480, pixel-aspect-ratio=(fraction)1/1, framerate=(fraction)30/1, interlace-mode=(string)progressive, colorimetry=(string)bt709
/GstPipeline:pipeline0/GstCapsFilter:capsfilter1.GstPad:sink: caps = video/x-h264, stream-format=(string)byte-stream, alignment=(string)au, level=(string)3, profile=(string)baseline, width=(int)640, height=(int)480, pixel-aspect-ratio=(fraction)1/1, framerate=(fraction)30/1, interlace-mode=(string)progressive, colorimetry=(string)bt709
Redistribute latency...
/GstPipeline:pipeline0/v4l2h264enc:v4l2h264enc0.GstPad:sink: caps = video/x-raw, width=(int)640, height=(int)480, framerate=(fraction)30/1, format=(string)YUY2, interlace-mode=(string)progressive, colorimetry=(string)bt709
/GstPipeline:pipeline0/GstVideoConvert:videoconvert0.GstPad:sink: caps = video/x-raw, format=(string)YUY2, width=(int)640, height=(int)480, colorimetry=(string)2:4:5:1, framerate=(fraction)30/1
/GstPipeline:pipeline0/GstCapsFilter:capsfilter0.GstPad:sink: caps = video/x-raw, format=(string)YUY2, width=(int)640, height=(int)480, colorimetry=(string)2:4:5:1, framerate=(fraction)30/1
^Chandling interrupt.
Interrupt: Stopping pipeline ...
Execution ended after 0:00:48.566840652
Setting pipeline to NULL ...
Freeing pipeline ...
********
*/
