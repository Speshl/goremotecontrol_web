package server

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Speshl/goremotecontrol_web/internal/gst"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

func (c *Connection) PlayTrack(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	log.Println("start playing client track")
	defer log.Println("done playing client track")
	// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
	go func() {
		ticker := time.NewTicker(time.Second * 3)
		for range ticker.C {
			if err := c.CTX.Err(); err != nil {
				return
			}
			rtcpSendErr := c.PeerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
			if rtcpSendErr != nil {
				log.Printf("error sending keyframe on ticker - %w\n", rtcpSendErr)
				return
			}
		}
	}()

	codecName := strings.Split(track.Codec().RTPCodecCapability.MimeType, "/")[1]
	fmt.Printf("Track has started, of type %d: %s \n", track.PayloadType(), codecName)
	pipeline := gst.CreateRecievePipeline(track.PayloadType(), strings.ToLower(codecName))
	pipeline.Start()
	buf := make([]byte, 1400)
	for {
		i, _, err := track.Read(buf)
		if err != nil {
			log.Printf("stopping client audio - error reading client audio track buffer - %w\n", err)
			return
		}
		//log.Printf("Pushing %d bytes to pipeline", i)
		pipeline.Push(buf[:i])
	}
}
