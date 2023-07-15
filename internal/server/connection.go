package server

import (
	"context"
	"fmt"
	"log"

	"github.com/Speshl/goremotecontrol_web/internal/carcam"
	socketio "github.com/googollee/go-socket.io"
	"github.com/pion/webrtc/v3"
)

type Connection struct {
	ID             string
	Socket         socketio.Conn
	PeerConnection *webrtc.PeerConnection
	Cancel         context.CancelFunc
	CTX            context.Context
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
		Socket:         socketConn,
		PeerConnection: peerConnection,
		Cancel:         cancelCTX,
		CTX:            ctx,
	}
	return conn, nil
}

func (c *Connection) Disconnect() {
	c.Cancel()
	c.PeerConnection.Close()
}

func (c *Connection) RegisterHandlers(car *carcam.CarCam) {
	err := c.addTracks(car.AudioTracks, car.VideoTracks)
	if err != nil {
		log.Printf("failed to add tracks: %w\n", err)
		return
	}

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

	encodedAnswer, err := encode(c.PeerConnection.LocalDescription())
	if err != nil {
		log.Printf("Failed encoding answer: %s", err.Error())
		return
	}
	c.Socket.Emit("answer", encodedAnswer)
}

func (c *Connection) addTracks(audioTracks []*webrtc.TrackLocalStaticSample, videoTracks []*webrtc.TrackLocalStaticSample) error {
	for _, audioTrack := range audioTracks {
		_, err := c.PeerConnection.AddTrack(audioTrack)
		if err != nil {
			return fmt.Errorf("error adding audio track: %w", err)
		}
	}

	for _, videoTrack := range videoTracks {
		_, err := c.PeerConnection.AddTrack(videoTrack)
		if err != nil {
			return fmt.Errorf("error adding video track: %w", err)
		}
	}
	return nil
}
