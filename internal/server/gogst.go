// This example shows how to use the appsink element.
package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
)

func createPipeline() (*gst.Pipeline, error) {
	log.Println("Creating Pipeline")
	gst.Init(nil)

	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, err
	}

	src, err := gst.NewElement("videotestsrc")
	if err != nil {
		return nil, err
	}

	sink, err := app.NewAppSink()
	if err != nil {
		return nil, err
	}

	pipeline.AddMany(src, sink.Element)
	src.Link(sink.Element)

	// Tell the appsink what format we want. It will then be the audiotestsrc's job to
	// provide the format we request.
	// This can be set after linking the two objects, because format negotiation between
	// both elements will happen during pre-rolling of the pipeline.
	sink.SetCaps(gst.NewCapsFromString(
		"audio/x-raw, format=S16LE, layout=interleaved, channels=1",
	))

	// Getting data out of the appsink is done by setting callbacks on it.
	// The appsink will then call those handlers, as soon as data is available.
	sink.SetCallbacks(&app.SinkCallbacks{
		// Add a "new-sample" callback
		NewSampleFunc: func(sink *app.Sink) gst.FlowReturn {
			log.Println("Got Sample")
			// Pull the sample that triggered this callback
			sample := sink.PullSample()
			if sample == nil {
				log.Println("gst flow eos")
				return gst.FlowEOS
			}

			// Retrieve the buffer from the sample
			buffer := sample.GetBuffer()
			if buffer == nil {
				log.Println("gst flow error")
				return gst.FlowError
			}

			// At this point, buffer is only a reference to an existing memory region somewhere.
			// When we want to access its content, we have to map it while requesting the required
			// mode of access (read, read/write).
			//
			// We also know what format to expect because we set it with the caps. So we convert
			// the map directly to signed 16-bit little-endian integers.
			//samples := buffer.Map(gst.MapRead).AsInt16LESlice()
			sampleBytes := buffer.Map(gst.MapRead).Bytes()
			defer buffer.Unmap()
			log.Printf("got %d bytes from sample: %d\n", len(sampleBytes))

			return gst.FlowOK
		},
	})

	return pipeline, nil
}

func handleMessage(msg *gst.Message) error {

	switch msg.Type() {
	case gst.MessageEOS:
		return app.ErrEOS
	case gst.MessageError:
		return msg.ParseError()
	}

	return nil
}

func mainLoop(ctx context.Context, pipeline *gst.Pipeline) error {
	log.Println("Starting main loop")
	// Start the pipeline
	pipeline.SetState(gst.StatePlaying)

	log.Println("Get pipeline bus")
	// Retrieve the bus from the pipeline
	bus := pipeline.GetPipelineBus()

	// Loop over messsages from the pipeline
	for {
		select {
		case <-ctx.Done():
			log.Println("Sending EOS Event")
			pipeline.SendEvent(gst.NewEOSEvent())
			return ctx.Err()
		default:
			msg := bus.TimedPop(time.Duration(-1))
			if msg == nil {
				return nil
			}
			if err := handleMessage(msg); err != nil {
				return err
			}
		}
	}
}

func StartGoGST(ctx context.Context) {
	Run(func() error {
		var err error
		var pipeline *gst.Pipeline
		if pipeline, err = createPipeline(); err != nil {
			return err
		}
		return mainLoop(ctx, pipeline)
	})
}

// Run is used to wrap the given function in a main loop and print any error
func Run(f func() error) {
	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

	go func() {
		if err := f(); err != nil {
			fmt.Println("ERROR!", err)
		}
		mainLoop.Quit()
	}()

	mainLoop.Run()
}
