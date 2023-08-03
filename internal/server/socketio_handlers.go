package server

import (
	"fmt"
	"log"

	socketio "github.com/googollee/go-socket.io"
	"github.com/pion/webrtc/v3"
)

func (s *Server) RegisterSocketIOHandlers() {
	s.socketio.OnConnect("/", s.onConnect)

	s.socketio.OnEvent("/", "offer", s.onOffer)

	s.socketio.OnEvent("/", "candidate", s.onICECandidate)

	s.socketio.OnEvent("/", "command", s.onCommand)

	s.socketio.OnDisconnect("/", s.OnDisconnect)

	s.socketio.OnError("/", s.onError)
}

func (s *Server) onConnect(socketConn socketio.Conn) error {
	log.Printf("Connected: %s\n", socketConn.ID())
	id := socketConn.ID()
	// Create a new Client for the connected socket
	conn, err := s.NewClientConn(socketConn)
	if err != nil {
		return fmt.Errorf("failed creating new client: %w", err)
	}

	s.connectionsLock.Lock()
	s.connections[id] = conn
	s.connectionsLock.Unlock()

	return nil
}

func (s *Server) onOffer(socketConn socketio.Conn, msg string) {
	log.Println("Offer Recieved From Connection:", socketConn.ID())
	//Send client answer to client's SDP answer channel
	s.connectionsLock.RLock()
	connection, ok := s.connections[socketConn.ID()]
	s.connectionsLock.RUnlock()
	if ok {
		offer := webrtc.SessionDescription{}
		err := decode(msg, &offer)
		if err != nil {
			log.Printf("Offer from %s failed unmarshaling: %s\n", socketConn.ID(), string(msg))
			return
		}
		go connection.ProcessOffer(offer)
	} else {
		log.Printf("got offer from unknown client: %s", socketConn.ID())
	}
}

func (s *Server) onICECandidate(socketConn socketio.Conn, msg []byte) {
	log.Println("candidate recieved from client")
}

func (s *Server) onCommand(socketConn socketio.Conn, msg []byte) {
	s.commandParser(msg)
}

func (s *Server) OnDisconnect(socketConn socketio.Conn, reason string) {
	log.Printf("connection disconnected (%s): %s\n", reason, socketConn.ID())
	s.RemoveClient(socketConn.ID())
}

func (s *Server) onError(socketConn socketio.Conn, err error) {
	log.Printf("connection %s error: %s\n", socketConn.ID(), err.Error())
}

func (s *Server) commandParser(msg []byte) {
	if len(msg) != 5 {
		log.Println("error: command is incorrect length")
	}

	s.commandChannel <- msg[0:4] //first 4 bytes go to carCommand

	//5th byte is a sound signal
	log.Printf("sound byte %d\n", msg[4])
	switch msg[4] {
	case 0:
		break
	case 1:
		s.speakerChannel <- "affirmative"
	case 2:
		s.speakerChannel <- "negative"
	case 3:
		s.speakerChannel <- "aggressive"
	case 4:
		s.speakerChannel <- "sorry"
	default:
		log.Println("error: invalid sound command")
	}
}
