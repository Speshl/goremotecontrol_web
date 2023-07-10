package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	"github.com/pion/webrtc/v3"
)

type Server struct {
	socketio    *socketio.Server
	clients     map[string]*Client
	clientsLock sync.RWMutex
}

type Client struct {
	ID             string
	Socket         socketio.Conn
	PeerConnection *webrtc.PeerConnection
	//AnswerChannel  chan webrtc.SessionDescription
	OfferChannel chan webrtc.SessionDescription
	Cancel       context.CancelFunc
	CTX          context.Context
	//RemoteCandidates []webrtc.ICECandidateInit
}

var allowOriginFunc = func(r *http.Request) bool {
	return true
}

func NewSocketServer() *Server {
	socketioServer := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: allowOriginFunc,
			},
			&websocket.Transport{
				CheckOrigin: allowOriginFunc,
			},
		},
	})

	return &Server{
		socketio: socketioServer,
		clients:  make(map[string]*Client),
	}
}

func (s *Server) RegisterHandlers() {
	s.socketio.OnConnect("/", func(socketConn socketio.Conn) error {
		log.Printf("Connected: %s\n", socketConn.ID())
		id := socketConn.ID()
		// Create a new Client for the connected socket
		client, err := s.NewClient(socketConn)
		if err != nil {
			return fmt.Errorf("failed creating new client: %w", err)
		}

		go client.handleOfferChannel()

		s.clientsLock.Lock()
		s.clients[id] = client
		s.clientsLock.Unlock()

		return nil
	})

	s.socketio.OnEvent("/", "offer", func(socketConn socketio.Conn, msg string) {
		log.Println("Client offered:", socketConn.ID())
		//Send client answer to client's SDP answer channel
		s.clientsLock.RLock()
		connection, ok := s.clients[socketConn.ID()]
		s.clientsLock.RUnlock()
		if ok {
			offer := webrtc.SessionDescription{}
			err := decode(msg, &offer)
			if err != nil {
				log.Printf("Offer from %s failed unmarshaling: %s\n", socketConn.ID(), string(msg))
				return
			}
			connection.OfferChannel <- offer
		} else {
			log.Printf("Got offer from unknown client: %s", socketConn.ID())
		}
	})

	// s.socketio.OnEvent("/", "answer", func(socketConn socketio.Conn, msg string) {
	// 	log.Println("Client answered:", socketConn.ID())
	// 	//Send client answer to client's SDP answer channel
	// 	s.clientsLock.RLock()
	// 	connection, ok := s.clients[socketConn.ID()]
	// 	s.clientsLock.RUnlock()
	// 	if ok {
	// 		answer := webrtc.SessionDescription{
	// 			Type: webrtc.SDPTypeAnswer,
	// 			SDP:  msg,
	// 		}

	// 		connection.AnswerChannel <- answer
	// 	}
	// })

	s.socketio.OnEvent("/", "candidate", func(socketConn socketio.Conn, msg []byte) {
		//socketConn.SetContext(msg)
		//log.Printf("candidate: %+v\n", msg)
		log.Println("Candidate recieved from clinet")
	})

	s.socketio.OnEvent("/", "command", func(socketConn socketio.Conn, msg []byte) {
		//socketConn.SetContext(msg)
		log.Printf("command: %+v\n", msg)
	})

	s.socketio.OnDisconnect("/", func(socketConn socketio.Conn, reason string) {
		log.Println("Client disconnected:", socketConn.ID())
		s.RemoveClient(socketConn.ID())
	})

	s.socketio.OnError("/", func(s socketio.Conn, e error) {
		log.Println("socketio error:", e)
	})
}

func (s *Server) RemoveClient(id string) {
	log.Printf("Remove Client %s\n", id)
	s.clientsLock.Lock()
	client, ok := s.clients[id]
	if ok {
		client.PeerConnection.Close()
		delete(s.clients, id)
		close(client.OfferChannel)
		client.Cancel()
	}
	s.clientsLock.Unlock()
}

func (s *Server) Close() error {
	return s.socketio.Close()
}

func (s *Server) Serve() error {
	return s.socketio.Serve()
}

func (s *Server) GetHandler() *socketio.Server {
	return s.socketio
}

// Encode encodes the input in base64
func encode(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// Decode decodes the input from base64
func decode(in string, obj interface{}) error {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, obj)
	if err != nil {
		return err
	}
	return nil
}
