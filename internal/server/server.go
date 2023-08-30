package server

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Speshl/goremotecontrol_web/internal/carcommand"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	"github.com/pion/webrtc/v3"
)

type ClientAudioTrackPlayer func(*webrtc.TrackRemote, *webrtc.RTPReceiver)

type Server struct {
	carAudioTrack    *webrtc.TrackLocalStaticSample
	carVideoTrack    *webrtc.TrackLocalStaticSample
	commandChannel   chan carcommand.CommandGroup
	memeSoundChannel chan string

	clientAudioTrackPlayer ClientAudioTrackPlayer

	socketio        *socketio.Server
	connections     map[string]*Connection
	connectionsLock sync.RWMutex

	config SocketServerConfig
}

type SocketServerConfig struct {
	SilentConnects bool
}

var allowOriginFunc = func(r *http.Request) bool {
	return true
}

func NewSocketServer(cfg SocketServerConfig, audioTrack *webrtc.TrackLocalStaticSample, videoTrack *webrtc.TrackLocalStaticSample, commandChannel chan carcommand.CommandGroup, memeSoundChannel chan string, audioPlayer ClientAudioTrackPlayer) *Server {
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
		socketio:    socketioServer,
		connections: make(map[string]*Connection),

		memeSoundChannel:       memeSoundChannel,
		commandChannel:         commandChannel,
		carAudioTrack:          audioTrack,
		carVideoTrack:          videoTrack,
		clientAudioTrackPlayer: audioPlayer,
		config:                 cfg,
	}
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

func (s *Server) NewClientConn(socketConn socketio.Conn) (*Connection, error) {
	clientConn, err := NewConnection(socketConn, s.clientAudioTrackPlayer)
	if err != nil {
		return nil, err
	}

	err = clientConn.RegisterHandlers(s.carAudioTrack, s.carVideoTrack, s.memeSoundChannel)
	if err != nil {
		return nil, err
	}

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	clientConn.PeerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("Peer Connection State has changed: %s\n", state.String())
		if state == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			log.Println("Peer Connection has gone to failed")
			s.RemoveClient(socketConn.ID())
		}
		if state == webrtc.PeerConnectionStateConnecting {
			s.memeSoundChannel <- "client_connected"
		} else if state == webrtc.PeerConnectionStateClosed {
			time.Sleep(5 * time.Second)
			s.memeSoundChannel <- "client_disconnected"
		}
	})

	return clientConn, nil
}

func (s *Server) RemoveClient(id string) {
	log.Printf("Removing Client: %s\n", id)
	s.connectionsLock.Lock()
	client, ok := s.connections[id]
	if ok {
		client.Disconnect()
		delete(s.connections, id)
		log.Printf("Client Removed: %s\n", id)
	}
	s.connectionsLock.Unlock()
}
