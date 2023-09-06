// Package gst provides an easy API to create an appsrc pipeline
package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "gst.h"

*/
import "C"
import (
	"fmt"
	"log"
	"strings"
	"unsafe"

	"github.com/pion/webrtc/v3"
)

// StartMainLoop starts GLib's main loop
// It needs to be called from the process' main thread
// Because many gstreamer plugins require access to the main thread
// See: https://golang.org/pkg/runtime/#LockOSThread
func StartMainRecieveLoop() {
	C.gstreamer_receive_start_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type RecievePipeline struct {
	Pipeline *C.GstElement
}

// CreatePipeline creates a GStreamer Pipeline
func CreateRecievePipeline(payloadType webrtc.PayloadType, codecName string, device string, volume string) *RecievePipeline {
	pipelineStr := "appsrc format=time is-live=true do-timestamp=true name=src ! application/x-rtp"
	switch strings.ToLower(codecName) {
	// case "vp8":
	// 	pipelineStr += fmt.Sprintf(", payload=%d, encoding-name=VP8-DRAFT-IETF-01 ! rtpvp8depay ! decodebin ! autovideosink", payloadType)
	case "opus":
		pipelineStr += fmt.Sprintf(", payload=%d, encoding-name=OPUS ! rtpopusdepay ! decodebin ! pulsesink device=%s volume=%s", payloadType, device, volume)
	// case "vp9":
	// 	pipelineStr += " ! rtpvp9depay ! decodebin ! autovideosink"
	// case "h264":
	// 	pipelineStr += " ! rtph264depay ! decodebin ! autovideosink"
	// case "g722":
	// 	pipelineStr += " clock-rate=8000 ! rtpg722depay ! decodebin ! pulsesink device=1"
	default:
		panic("Unhandled codec " + codecName)
	}
	log.Printf("client audio pipeline: %s\n", pipelineStr)
	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))
	return &RecievePipeline{Pipeline: C.gstreamer_receive_create_pipeline(pipelineStrUnsafe)}
}

// Start starts the GStreamer Pipeline
func (p *RecievePipeline) Start() {
	C.gstreamer_receive_start_pipeline(p.Pipeline)
}

// Stop stops the GStreamer Pipeline
func (p *RecievePipeline) Stop() {
	C.gstreamer_receive_stop_pipeline(p.Pipeline)
}

// Push pushes a buffer on the appsrc of the GStreamer Pipeline
func (p *RecievePipeline) Push(buffer []byte) {
	b := C.CBytes(buffer)
	defer C.free(b)
	C.gstreamer_receive_push_buffer(p.Pipeline, b, C.int(len(buffer)))
}
