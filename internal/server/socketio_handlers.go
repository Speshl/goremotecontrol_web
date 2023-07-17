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
	//s.carCommand.CommandChannel <- msg
}

func (s *Server) OnDisconnect(socketConn socketio.Conn, reason string) {
	log.Printf("connection disconnected (%s): %s\n", reason, socketConn.ID())
	s.RemoveClient(socketConn.ID())
}

func (s *Server) onError(socketConn socketio.Conn, err error) {
	log.Printf("connection %s error: %s\n", socketConn.ID(), err.Error())
}
