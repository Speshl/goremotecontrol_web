package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	socketio "github.com/googollee/go-socket.io"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"
	"github.com/pion/webrtc/v3/pkg/media/oggreader"

	x264 "github.com/gen2brain/x264-go"
)

var opts = &x264.Options{
	Width:     640,
	Height:    480,
	FrameRate: 25,
	Tune:      "zerolatency",
	Preset:    "ultrafast",
	Profile:   "baseline",
	LogLevel:  x264.LogDebug,
}

func (s *Server) NewClient(socketConn socketio.Conn) (*Client, error) {
	id := socketConn.ID()
	log.Printf("Creating Client %s\n", id)
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Println("Failed to create PeerConnection:", err)
		return nil, err
	}

	ctx, cancelCTX := context.WithCancel(context.Background())

	client := &Client{
		ID: id,
		//AnswerChannel:  make(chan webrtc.SessionDescription),
		OfferChannel:   make(chan webrtc.SessionDescription),
		Socket:         socketConn,
		PeerConnection: peerConnection,
		Cancel:         cancelCTX,
		CTX:            ctx,
		//RemoteCandidates: []webrtc.ICECandidateInit{},
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("Connection State has changed: %s\n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			log.Println("Peer ICEConnectionStateConnected")
		}
	})

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("Peer Connection State has changed: %s\n", state.String())
		if state == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			log.Println("Peer Connection has gone to failed")
			s.RemoveClient(socketConn.ID())
		}
	})

	// Handle ICE candidate messages from the client
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			log.Println("Sending ICE candidate to client:", candidate)
			encodedCandidate, err := encode(candidate.ToJSON())
			if err != nil {
				log.Printf("Error encoding candidate: %s", err.Error())
			}
			client.Socket.Emit("candidate", encodedCandidate)
		}
	})

	// Add the data channel to the peer connection
	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		log.Println("Failed to create data channel:", err)
		return nil, err
	}

	// Handle data channel messages
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		log.Println("Received data channel message:", string(msg.Data))
	})
	return client, nil
}

func (c *Client) handleOfferChannel() {
	for { //Add ctx cancel check back in
		select {
		case offer, ok := <-c.OfferChannel:
			if !ok {
				log.Println("Offer channel closed")
				return
			}
			// Process the received offer
			log.Printf("Received Offer size: %d\n", len(offer.SDP))
			//log.Printf("Offer: %s\n\n", offer.SDP)

			// Set the received offer as the remote description
			err := c.PeerConnection.SetRemoteDescription(offer)
			if err != nil {
				log.Println("Failed to set remote description:", err)
				return
			}

			// err = c.PlayTempAudio(c.CTX)
			// if err != nil {
			// 	log.Printf("Failed playing audio: %s", err.Error())
			// 	return
			// }

			err = c.TempStreamVideo(c.CTX)
			if err != nil {
				log.Printf("Failed playing video: %s", err.Error())
				return
			}

			// Create answer
			answer, err := c.PeerConnection.CreateAnswer(nil)
			if err != nil {
				log.Println("Failed to create answer:", err)
				return
			}

			// Create channel that is blocked until ICE Gathering is complete
			//gatherComplete := webrtc.GatheringCompletePromise(c.PeerConnection)

			// Sets the LocalDescription, and starts our UDP listeners
			err = c.PeerConnection.SetLocalDescription(answer)
			if err != nil {
				log.Println("Failed to set local description:", err)
				return
			}

			// Block until ICE Gathering is complete, disabling trickle ICE
			// we do this because we only can exchange one signaling message
			// in a production application you should exchange ICE Candidates via OnICECandidate
			// log.Println("Waiting for ICE Gathering")
			// <-gatherComplete
			// log.Println("ICE Gathering Complete")

			log.Printf("Answer Size: %d\n\n", len(c.PeerConnection.LocalDescription().SDP))
			//log.Printf("Answer: %s\n\n", c.PeerConnection.LocalDescription().SDP)

			encodedAnswer, err := encode(c.PeerConnection.LocalDescription())
			if err != nil {
				log.Printf("Failed encoding answer: %s", err.Error())
				return
			}
			c.Socket.Emit("answer", encodedAnswer) //TODO: Is this how to send the answer?
		}
	}
}

func (c *Client) PlayTempAudio(ctx context.Context) error {
	log.Println("Start Temp Audio Player")
	defer log.Println("End Temp Audio Player")

	filePath := "./test_data/output.ogg"
	oggPageDuration := time.Millisecond * 20

	_, err := os.Stat(filePath)
	haveAudioFile := !os.IsNotExist(err)
	if !haveAudioFile {
		return err
	}

	// Create a audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
	if err != nil {
		return err
	}

	rtpSender, err := c.PeerConnection.AddTrack(audioTrack)
	if err != nil {
		return err
	}

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		defer log.Println("Done doing whatever this does with Audio")
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				log.Printf("RTCP Error: %s", rtcpErr.Error())
				return
			}

			_, ok := <-ctx.Done()
			if !ok {
				log.Println("Context closed, stopping rtpSender Reader")
				return
			}
		}
	}()

	go func() {
		log.Println("Start Playing Audio")
		defer log.Println("Done playing Audio")
		// Open a OGG file and start reading using our OGGReader
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("Error opening audio file: %s\n", err.Error())
			return
		}

		// Open on oggfile in non-checksum mode.
		ogg, _, err := oggreader.NewWith(file)
		if err != nil {
			log.Printf("Error reading audio file: %s\n", err.Error())
			return
		}

		// Keep track of last granule, the difference is the amount of samples in the buffer
		var lastGranule uint64

		// It is important to use a time.Ticker instead of time.Sleep because
		// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
		// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
		ticker := time.NewTicker(oggPageDuration)
		time.Sleep(2000)
		for {
			select {
			case _, ok := <-ctx.Done():
				if !ok {
					log.Println("Context closed, stopping rtpSender Reader")
					return
				}
			case <-ticker.C:
				pageData, pageHeader, err := ogg.ParseNextPage()
				if errors.Is(err, io.EOF) {
					log.Println("All audio pages parsed and sent")
					return
				}

				if err != nil {
					log.Printf("Error parsing and sending audio pages: %s\n", err)
					return
				}

				// The amount of samples is the difference between the last and current timestamp
				sampleCount := float64(pageHeader.GranulePosition - lastGranule)
				lastGranule = pageHeader.GranulePosition
				sampleDuration := time.Duration((sampleCount/48000)*1000) * time.Millisecond

				err = audioTrack.WriteSample(media.Sample{Data: pageData, Duration: sampleDuration})
				if err != nil {
					log.Printf("Error parsing and sending audio pages: %s\n", err)
					return
				}
			}
		}
	}()
	return nil
}

func (c *Client) PlayTempCam(ctx context.Context) error {
	return nil
}

func (c *Client) TempStreamVideo(ctx context.Context) error {
	log.Println("Start streaming video")
	frameDuration := time.Millisecond * 40

	videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if videoTrackErr != nil {
		panic(videoTrackErr)
	}

	rtpSender, videoTrackErr := c.PeerConnection.AddTrack(videoTrack)
	if videoTrackErr != nil {
		panic(videoTrackErr)
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	go func() {
		defer log.Println("Done streaming video")
		file, err := os.Open("screen.264") //screen.264 output.h264
		if err != nil {
			log.Printf("failed opening video file: %w", err)
			return
		}
		defer file.Close()

		fileReader, err := h264reader.NewReader(file)
		if err != nil {
			fmt.Errorf("failed opening with h264 reader: %w", err)
			return
		}

		ticker := time.NewTicker(frameDuration)
		for ; true; <-ticker.C {
			log.Println("Sending frame")
			nal, h264Err := fileReader.NextNAL()
			if h264Err == io.EOF {
				fmt.Printf("All video frames parsed and sent")
				return
			}
			if h264Err != nil {
				log.Printf("error reading next nal: %s\n", err.Error())
				return
			}

			if h264Err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: frameDuration}); h264Err != nil {
				log.Printf("error writing sample: %s\n", err.Error())
				return
			}
		}
	}()
	return nil
}

// func TempRecordCam() (err error) {
// 	deviceID := 0
// 	webcam, err := gocv.VideoCaptureDevice(deviceID)
// 	if err != nil {
// 		log.Printf("Failing opening video capture device: %s\n", err.Error())
// 		return err
// 	}

// 	file, err := os.Create("screen.264")
// 	if err != nil {
// 		return fmt.Errorf("Error creating video file: %w", err)
// 	}

// 	enc, err := x264.NewEncoder(file, opts)
// 	if err != nil {
// 		return fmt.Errorf("Error creating encoder: %w", err)
// 	}

// 	frameTicker := time.NewTicker(time.Second / time.Duration(25))

// 	defer func() {
// 		err = enc.Flush()
// 		if err != nil {
// 			err = fmt.Errorf("Error flushing encoder: %w", err)
// 			return
// 		}

// 		file.Close()
// 		if err != nil {
// 			err = fmt.Errorf("Error flushing encoder: %w", err)
// 		}

// 	}()

// 	frameCounter := 0
// 	for range frameTicker.C {
// 		if frameCounter > 500 {
// 			frameTicker.Stop()
// 			return nil
// 		}

// 		matImg := gocv.NewMat()
// 		ok := webcam.Read(&matImg)
// 		if !ok {
// 			return fmt.Errorf("error reading from video devide %d\n", deviceID)
// 		}

// 		img, err := matImg.ToImage()
// 		if err != nil {
// 			return fmt.Errorf("error converting from gocv mat to image: %w", err)
// 		}

// 		// img := x264.NewYCbCr(image.Rect(0, 0, opts.Width, opts.Height))
// 		// draw.Draw(img, img.Bounds(), image.Black, image.ZP, draw.Src)
// 		// img.Set(frameCounter, opts.Height/2, color.RGBA{255, 0, 0, 255})

// 		log.Println("Encoding frame %d\n", frameCounter)
// 		frameCounter++
// 		err = enc.Encode(img)
// 		if err != nil {
// 			return fmt.Errorf("Error encoding frame: %w", err)
// 		}
// 	}
// 	return nil
// }

// func TempRecordPainted() (err error) {
// 	file, err := os.Create("screen.264")
// 	if err != nil {
// 		return fmt.Errorf("Error creating video file: %w", err)
// 	}

// 	enc, err := x264.NewEncoder(file, opts)
// 	if err != nil {
// 		return fmt.Errorf("Error creating encoder: %w", err)
// 	}

// 	frameTicker := time.NewTicker(time.Second / time.Duration(25))
// 	stopTime := time.Now().Add(time.Second * 5)

// 	defer func() {
// 		err = enc.Flush()
// 		if err != nil {
// 			err = fmt.Errorf("Error flushing encoder: %w", err)
// 			return
// 		}

// 		file.Close()
// 		if err != nil {
// 			err = fmt.Errorf("Error flushing encoder: %w", err)
// 		}

// 	}()

// 	frameCounter := 0
// 	for range frameTicker.C {
// 		if frameCounter > 500 {
// 			frameTicker.Stop()
// 			return nil
// 		}
// 		img := x264.NewYCbCr(image.Rect(0, 0, opts.Width, opts.Height))
// 		draw.Draw(img, img.Bounds(), image.Black, image.ZP, draw.Src)
// 		img.Set(frameCounter, opts.Height/2, color.RGBA{255, 0, 0, 255})

// 		log.Println("Encoding frame %d\n", frameCounter)
// 		frameCounter++
// 		err = enc.Encode(img)
// 		if err != nil {
// 			return fmt.Errorf("Error encoding frame: %w", err)
// 		}
// 	}
// 	return nil
// }

// func PlayReadWebCam() error {
// 	log.Printf("Start Reading Webcam")
// 	defer log.Printf("Done Reading Webcam")

// 	deviceID := 0
// 	webcam, err := gocv.VideoCaptureDevice(deviceID)
// 	if err != nil {
// 		log.Printf("Failing opening video capture device: %s\n", err.Error())
// 		return err
// 	}

// 	img := gocv.NewMat()
// 	ok := webcam.Read(&img)
// 	if !ok {
// 		return fmt.Errorf("error reading from video devide %d\n", deviceID)
// 	}

// 	size := img.Size()
// 	log.Printf("Frame Empty: %t Frame Size - 0: %d 1: %d\n", img.Empty(), size[0], size[1])
// 	return nil
// }
