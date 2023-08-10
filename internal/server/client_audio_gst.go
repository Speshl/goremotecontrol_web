package server

// type AudioBuffer struct {
// 	lock   sync.RWMutex
// 	buffer []byte
// }

// func (c *Connection) createClientAudioPipeline(track *webrtc.TrackRemote) (*gst.Pipeline, error) {
// 	// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
// 	go func() { //No clue why I need this?
// 		ticker := time.NewTicker(time.Second * 3)
// 		for range ticker.C {
// 			log.Println("Ticking")
// 			errSend := c.PeerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
// 			if errSend != nil {
// 				log.Printf("pli keyframe thingy error - %s\n", errSend.Error())
// 			}
// 		}
// 	}()

// 	log.Printf("Got track from client: %+v\n\n", track)

// 	codecName := strings.Split(track.Codec().RTPCodecCapability.MimeType, "/")[1]
// 	log.Printf("Track has started, of type %d: %s \n\n", track.PayloadType(), codecName)

// 	gst.Init(nil)

// 	log.Println("building pipeline")
// 	//pipelineString := fmt.Sprintf("appsrc format=time is-live=true do-timestamp=true name=src ! application/x-rtp, payload=%d, encoding-name=OPUS ! rtpopusdepay ! decodebin ! pulsesink device=1", track.PayloadType())
// 	pipeline, err := gst.NewPipelineFromString("appsrc format=bytes is-live=true do-timestamp=true name=src ! application/x-rtp, payload=111, encoding-name=OPUS ! queue ! rtpopusdepay ! decodebin ! pulsesink device=1")
// 	if err != nil {
// 		return nil, fmt.Errorf("error creating client audio pipeline - %s\n", err.Error())
// 	}

// 	log.Println("getting elements")
// 	srcElement, err := pipeline.GetElementByName("src")
// 	if err != nil {
// 		return nil, fmt.Errorf("error getting client audio pipeline elements - %s\n", err.Error())
// 	}
// 	/*pipeline, err := gst.NewPipeline("")
// 	if err != nil {
// 		return nil, fmt.Errorf("error creating client audio pipeline - %s\n", err.Error())
// 	}

// 	elems, err := gst.NewElementMany("appsrc", "rtpopusdepay", "decodebin", "pulsesink")
// 	if err != nil {
// 		return nil, fmt.Errorf("error adding client audio elements to pipeline - %s\n", err.Error())
// 	}

// 	val, err := glib.ValueInit(glib.TYPE_ENUM)
// 	if err != nil {
// 		return nil, err
// 	}
// 	val.SetEnum(3)

// 	log.Printf("Val: %+v\n", val)

// 	err = elems[0].SetPropertyValue("format", val)
// 	if err != nil {
// 		return nil, fmt.Errorf("error setting format property - %s\n", err.Error())
// 	}

// 	// err = elems[0].SetProperty("is-live", true)
// 	// if err != nil {
// 	// 	return nil, fmt.Errorf("error setting is-live property - %s\n", err.Error())
// 	// }

// 	// err = elems[0].SetProperty("do-timestamp", true)
// 	// if err != nil {
// 	// 	return nil, fmt.Errorf("error setting do-timestamp property - %s\n", err.Error())
// 	// }

// 	// err = elems[0].SetProperty("name", "src")
// 	// if err != nil {
// 	// 	return nil, fmt.Errorf("error setting name property - %s\n", err.Error())
// 	// }

// 	err = elems[3].SetProperty("device", "1") //The sound hat device id from            pacmd list-cards                    index: ?
// 	if err != nil {
// 		return nil, fmt.Errorf("error setting audio output device - %s\n", err.Error())
// 	}

// 	capsString := fmt.Sprintf("application/x-rtp, payload= %d, encoding-name=OPUS", track.PayloadType())
// 	srcCaps := gst.NewCapsFromString(capsString)
// 	//srcCaps := gst.NewAnyCaps()

// 	pipeline.AddMany(elems...)
// 	gst.ElementLinkMany(elems...)
// 	*/

// 	log.Println("setting source")
// 	src := app.SrcFromElement(srcElement)

// 	//capsString := fmt.Sprintf("application/x-rtp, payload= %d, encoding-name=OPUS", track.PayloadType())
// 	log.Println("build src caps")
// 	//srcCaps := gst.NewCapsFromString(capsString)

// 	log.Println("set source")
// 	//src.SetCaps(srcCaps)

// 	log.Println("Setting callback")

// 	audioBuffer := AudioBuffer{}

// 	go func() {
// 		for {
// 			buf := make([]byte, 1400)
// 			numRead, _, err := track.Read(buf)
// 			if err != nil {
// 				log.Printf("error reading from client audio buffer: %s\n", err.Error())
// 				return
// 			}
// 			audioBuffer.lock.Lock()
// 			audioBuffer.buffer = append(audioBuffer.buffer, buf[:numRead]...)
// 			audioBuffer.lock.Unlock()
// 		}

// 	}()

// 	src.SetCallbacks(&app.SourceCallbacks{
// 		NeedDataFunc: func(self *app.Source, _ uint) {
// 			log.Println("client audio needs more data")

// 			audioBuffer.lock.Lock()
// 			numRead := len(audioBuffer.buffer)
// 			for numRead <= 0 {
// 				audioBuffer.lock.Unlock()
// 				time.Sleep(2 * time.Millisecond)
// 				audioBuffer.lock.Lock()
// 				numRead = len(audioBuffer.buffer)
// 			}

// 			log.Printf("Got %d bytes from track\n", numRead)

// 			buffer := gst.NewBufferWithSize(int64(numRead))
// 			buffer.Map(gst.MapWrite).WriteData(audioBuffer.buffer[0:numRead]) //send all recieved bytes since last asked
// 			buffer.Unmap()

// 			audioBuffer.buffer = audioBuffer.buffer[:0]
// 			audioBuffer.lock.Unlock()

// 			// Push the buffer onto the pipeline.
// 			self.PushBuffer(buffer) //Only push the number of bytes read
// 			log.Println("client audio data send")
// 		},
// 	})
// 	log.Println("Returning pipeline")
// 	return pipeline, nil
// }

// func (c *Connection) handleMessage(msg *gst.Message) error {
// 	log.Printf("GST Message: %s\n", msg.String())

// 	switch msg.Type() {
// 	case gst.MessageEOS:
// 		return app.ErrEOS
// 	case gst.MessageError:
// 		return msg.ParseError()
// 	}

// 	return nil
// }

// func (c *Connection) mainLoop(ctx context.Context, pipeline *gst.Pipeline) error {
// 	log.Println("Starting main client audio loop")
// 	// Start the pipeline
// 	pipeline.SetState(gst.StatePlaying)

// 	log.Println("Get client audio pipeline bus")
// 	// Retrieve the bus from the pipeline
// 	bus := pipeline.GetPipelineBus()

// 	// Loop over messsages from the pipeline
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Println("sending client audio EOS event")
// 			pipeline.SendEvent(gst.NewEOSEvent())
// 			return ctx.Err()
// 		default:
// 			msg := bus.TimedPop(time.Duration(-1))
// 			if msg == nil {
// 				return nil
// 			}
// 			if err := c.handleMessage(msg); err != nil {
// 				return err
// 			}
// 		}
// 	}
// }

// func (c *Connection) StartClientAudio(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
// 	c.run(func() error {
// 		var err error
// 		var pipeline *gst.Pipeline
// 		if pipeline, err = c.createClientAudioPipeline(track); err != nil {
// 			return fmt.Errorf("error creating pipeline - %w", err)
// 		}
// 		return c.mainLoop(c.CTX, pipeline)
// 	})
// 	return
// }

// // Run is used to wrap the given function in a main loop and print any error
// func (c *Connection) run(f func() error) {
// 	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

// 	go func() {
// 		if err := f(); err != nil {
// 			log.Printf("client audio error: %s", err.Error())
// 		}
// 		mainLoop.Quit()
// 	}()

// 	mainLoop.Run()
// }
