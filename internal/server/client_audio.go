package server

import (
	"log"
	"strings"

	"github.com/pion/webrtc/v3"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
)

func (c *Connection) playClientMic(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
	// go func() { //No clue why I need this?
	// 	ticker := time.NewTicker(time.Second * 3)
	// 	for range ticker.C {
	// 		errSend := c.PeerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
	// 		if errSend != nil {
	// 			log.Printf("pli keyframe thingy error - %s\n", errSend.Error())
	// 		}
	// 	}
	// }()

	log.Printf("Got track from client: %+v\n", track)

	log.Printf("Track Type: ")
	codecName := strings.Split(track.Codec().RTPCodecCapability.MimeType, "/")[1]
	log.Printf("Track has started, of type %d: %s \n", track.PayloadType(), codecName)

	gst.Init(nil)

	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, err
	}

	elems, err := gst.NewElementMany("appsrc", "opusdec", "audioconvert", "pulsesink")
	if err != nil {
		return nil, err
	}

	elems[2].SetProperty("device", 1) //The sound hat device id from            pacmd list-cards                    index: ?

	pipeline.AddMany(elems...)
	gst.ElementLinkMany(elems...)

	src := app.SrcFromElement(elems[0])

	src.SetCaps("audio/opus,rate=48000,channels=2")

	src.SetCallbacks(&app.SourceCallbacks{
		NeedDataFunc: func(self *app.Source, _ uint) {

			// Create a buffer that can hold exactly one video RGBA frame.
			buffer := gst.NewBufferWithSize(1400)

			// For each frame we produce, we set the timestamp when it should be displayed
			// The autovideosink will use this information to display the frame at the right time.
			//buffer.SetPresentationTimestamp(time.Duration(i) * 500 * time.Millisecond)

			//TODO: Get Audio Samples
			buf := make([]byte, 1400)
			numRead, _, err := track.Read(buf)
			if err != nil {
				log.Printf("error reading from client audio buffer: %s\n", err.Error())
			}

			// At this point, buffer is only a reference to an existing memory region somewhere.
			// When we want to access its content, we have to map it while requesting the required
			// mode of access (read, read/write).
			// See: https://gstreamer.freedesktop.org/documentation/plugin-development/advanced/allocation.html
			//
			// There are convenience wrappers for building buffers directly from byte sequences as
			// well.
			buffer.Map(gst.MapWrite).WriteData(buf)
			buffer.Unmap()

			// Push the buffer onto the pipeline.
			self.PushBuffer(buffer[:numRead]) //Only push the number of bytes read
		},
	})
	//src.SetProperty("format", gst.FormatTime)

	// buf := make([]byte, 1400)
	// for {
	// 	_, _, readErr := track.Read(buf)
	// 	if readErr != nil {
	// 		log.Printf("error reading client audio track - %s\n", err.Error())
	// 	}
	// 	//pipeline.Push(buf[:i])
	// }
	return
}
