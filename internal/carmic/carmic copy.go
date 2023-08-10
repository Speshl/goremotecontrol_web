package carmic

// type CarMic struct {
// 	AudioTrack *webrtc.TrackLocalStaticSample
// 	options    MicOptions
// }

// type MicOptions struct {
// 	Name string
// }

// func NewCarMic(options MicOptions) (*CarMic, error) {
// 	// Create a audio track
// 	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
// 	if err != nil {
// 		return nil, fmt.Errorf("error creating audio track: %w", err)
// 	}

// 	return &CarMic{
// 		AudioTrack: audioTrack,
// 		options:    options,
// 	}, nil
// }

// func (c *CarMic) createPipeline() (*gst.Pipeline, error) {
// 	log.Println("Creating Pipeline")
// 	gst.Init(nil)

// 	pipeline, err := gst.NewPipeline("")
// 	if err != nil {
// 		return nil, err
// 	}

// 	src, err := gst.NewElement("pulsesrc")
// 	if err != nil {
// 		return nil, err
// 	}

// 	converter, err := gst.NewElement("audioconvert")
// 	if err != nil {
// 		return nil, err
// 	}

// 	opusenc, err := gst.NewElement("opusenc")
// 	if err != nil {
// 		return nil, err
// 	}

// 	sink, err := app.NewAppSink()
// 	if err != nil {
// 		return nil, err
// 	}

// 	pipeline.AddMany(src, converter, opusenc, sink.Element)
// 	src.Link(converter)
// 	converter.Link(opusenc)
// 	opusenc.Link(sink.Element)

// 	// Tell the appsink what format we want. It will then be the audiotestsrc's job to
// 	// provide the format we request.
// 	// This can be set after linking the two objects, because format negotiation between
// 	// both elements will happen during pre-rolling of the pipeline.
// 	// sink.SetCaps(gst.NewCapsFromString(
// 	// 	"audio/opus",
// 	// ))

// 	// Getting data out of the appsink is done by setting callbacks on it.
// 	// The appsink will then call those handlers, as soon as data is available.
// 	sink.SetCallbacks(&app.SinkCallbacks{
// 		// Add a "new-sample" callback
// 		NewSampleFunc: func(sink *app.Sink) gst.FlowReturn {
// 			//log.Println("Got Sample")
// 			// Pull the sample that triggered this callback
// 			sample := sink.PullSample()
// 			if sample == nil {
// 				log.Println("gst flow eos")
// 				return gst.FlowEOS
// 			}

// 			//log.Printf("Sample Caps: %s\n", sample.GetCaps().String())

// 			// Retrieve the buffer from the sample
// 			buffer := sample.GetBuffer()
// 			if buffer == nil {
// 				log.Println("gst flow error")
// 				return gst.FlowError
// 			}

// 			// At this point, buffer is only a reference to an existing memory region somewhere.
// 			// When we want to access its content, we have to map it while requesting the required
// 			// mode of access (read, read/write).
// 			sampleBytes := buffer.Map(gst.MapRead).Bytes()

// 			c.AudioTrack.WriteSample(media.Sample{
// 				Data:     sampleBytes,
// 				Duration: time.Millisecond * 20, //This is the default for opus
// 			})

// 			defer buffer.Unmap()
// 			//log.Printf("got %d bytes from sample\n", len(sampleBytes)

// 			return gst.FlowOK
// 		},
// 	})

// 	return pipeline, nil
// }

// func (c *CarMic) handleMessage(msg *gst.Message) error {
// 	log.Printf("GST Message: %s\n", msg.String())

// 	switch msg.Type() {
// 	case gst.MessageEOS:
// 		return app.ErrEOS
// 	case gst.MessageError:
// 		return msg.ParseError()
// 	}

// 	return nil
// }

// func (c *CarMic) mainLoop(ctx context.Context, pipeline *gst.Pipeline) error {
// 	log.Println("Starting main loop")
// 	// Start the pipeline
// 	pipeline.SetState(gst.StatePlaying)

// 	log.Println("Get pipeline bus")
// 	// Retrieve the bus from the pipeline
// 	bus := pipeline.GetPipelineBus()

// 	// Loop over messsages from the pipeline
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Println("Sending EOS Event")
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

// func (c *CarMic) Start(ctx context.Context) error {
// 	c.run(func() error {
// 		var err error
// 		var pipeline *gst.Pipeline
// 		if pipeline, err = c.createPipeline(); err != nil {
// 			return err
// 		}
// 		return c.mainLoop(ctx, pipeline)
// 	})
// 	return nil
// }

// // Run is used to wrap the given function in a main loop and print any error
// func (c *CarMic) run(f func() error) {
// 	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

// 	go func() {
// 		if err := f(); err != nil {
// 			log.Printf("carmic error: %s", err.Error())
// 		}
// 		mainLoop.Quit()
// 	}()

// 	mainLoop.Run()
// }
