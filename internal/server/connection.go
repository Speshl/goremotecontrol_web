package server

import (
	"context"
	"fmt"
	"log"

	"github.com/Speshl/goremotecontrol_web/internal/server/gst"
	socketio "github.com/googollee/go-socket.io"
	"github.com/pion/webrtc/v3"
)

type Connection struct {
	ID             string
	Socket         socketio.Conn
	PeerConnection *webrtc.PeerConnection
	Cancel         context.CancelFunc
	CTX            context.Context

	AudioTracks []*webrtc.TrackLocalStaticSample
	VideoTracks []*webrtc.TrackLocalStaticSample

	AudioPipeline *gst.Pipeline
	VideoPipeline *gst.Pipeline
}

func NewConnection(socketConn socketio.Conn) (*Connection, error) {
	log.Printf("Creating Client %s\n", socketConn.ID())

	ctx, cancelCTX := context.WithCancel(context.Background())

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create Peer Connection: %s\n", err)
	}

	conn := &Connection{
		ID:             socketConn.ID(),
		AudioTracks:    make([]*webrtc.TrackLocalStaticSample, 0),
		VideoTracks:    make([]*webrtc.TrackLocalStaticSample, 0),
		Socket:         socketConn,
		PeerConnection: peerConnection,
		Cancel:         cancelCTX,
		CTX:            ctx,
	}
	return conn, nil
}

func (c *Connection) Disconnect() {
	c.Cancel()
	c.startStreaming()
	c.PeerConnection.Close()
}

func (c *Connection) RegisterHandlers() {
	c.addTracks()

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	c.PeerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("Connection State has changed: %s\n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			log.Println("Peer ICEConnectionStateConnected")
		}
	})

	// Handle ICE candidate messages from the client
	c.PeerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			log.Println("Sending ICE candidate to client:", candidate)
			encodedCandidate, err := encode(candidate.ToJSON())
			if err != nil {
				log.Printf("Error encoding candidate: %s", err.Error())
			}
			c.Socket.Emit("candidate", encodedCandidate)
		}
	})

	// // Add the data channel to the peer connection
	// dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	// if err != nil {
	// 	log.Println("Failed to create data channel:", err)
	// 	return nil, err
	// }

	// // Handle data channel messages
	// dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
	// 	log.Println("Received data channel message:", string(msg.Data))
	// })

}

func (c *Connection) ProcessOffer(offer webrtc.SessionDescription) {
	log.Printf("Received Offer size: %d\n", len(offer.SDP))

	// Set the received offer as the remote description
	err := c.PeerConnection.SetRemoteDescription(offer)
	if err != nil {
		log.Printf("failed to set remote description: %s\n", err)
		return
	}

	err = c.addTracks()
	if err != nil {
		log.Printf("failed to add tracks: %w\n", err)
		return
	}

	// Create answer
	answer, err := c.PeerConnection.CreateAnswer(nil)
	if err != nil {
		log.Printf("Failed to create answer: %s\n", err)
		return
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = c.PeerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Println("Failed to set local description:", err)
		return
	}

	c.startStreaming()

	encodedAnswer, err := encode(c.PeerConnection.LocalDescription())
	if err != nil {
		log.Printf("Failed encoding answer: %s", err.Error())
		return
	}
	c.Socket.Emit("answer", encodedAnswer)
}

func (c *Connection) addTracks() error {
	// Create a audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		return fmt.Errorf("error creating audio track: %w", err)
	}
	_, err = c.PeerConnection.AddTrack(audioTrack)
	if err != nil {
		return fmt.Errorf("error adding audio track: %w", err)
	}

	c.AudioTracks = append(c.AudioTracks, audioTrack)

	// Create a video track
	firstVideoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion2")
	if err != nil {
		return fmt.Errorf("error creating first video track: %w", err)
	}
	_, err = c.PeerConnection.AddTrack(firstVideoTrack)
	if err != nil {
		return fmt.Errorf("error adding first video track: %w", err)
	}

	c.VideoTracks = append(c.VideoTracks, firstVideoTrack)

	// Create a second video track
	secondVideoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion3")
	if err != nil {
		return fmt.Errorf("error creating second video track: %w", err)
	}
	_, err = c.PeerConnection.AddTrack(secondVideoTrack)
	if err != nil {
		return fmt.Errorf("error adding second video track: %w", err)
	}

	c.VideoTracks = append(c.VideoTracks, secondVideoTrack)
	return nil
}

func (c *Connection) startStreaming() {
	fmt.Printf("connection %s starting streams...\n", c.Socket.ID())
	audioSrc := "audiotestsrc" //audiotestsrc
	c.AudioPipeline = gst.CreatePipeline("opus", []*webrtc.TrackLocalStaticSample{c.AudioTracks[0]}, audioSrc)
	c.AudioPipeline.Start()

	//libcamerasrc ! video/x-raw, width=640, height=480, framerate=30/1 ! videoconvert
	//autovideosrc ! video/x-raw, width=320, height=240 ! videoconvert ! queue
	videoSrc := "libcamerasrc ! video/x-raw, width=640, height=480, framerate=30/1 ! videoconvert ! queue" //autovideosrc videotestsrc
	c.VideoPipeline = gst.CreatePipeline("vp8", []*webrtc.TrackLocalStaticSample{c.VideoTracks[0], c.VideoTracks[1]}, videoSrc)
	c.VideoPipeline.Start()
}

func (c *Connection) stopStreaming() {
	fmt.Printf("connection %s stopping streams...\n", c.Socket.ID())
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
