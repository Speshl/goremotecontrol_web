package carcam

import (
	"bytes"
	"context"
	"io"
	"log"
	"os/exec"
	"strconv"
)

var readBufferSize = 4096
var bufferSizeKB = 256
var nalSeparator = []byte{0, 0, 0, 1} //NAL break

// type CameraOptions struct {
// 	Width               int
// 	Height              int
// 	Fps                 int
// 	HorizontalFlip      bool
// 	VerticalFlip        bool
// 	Rotation            int
// 	UseLibcamera        bool // Set to true to enable libcamera, otherwise use legacy raspivid stack
// 	AutoDetectLibCamera bool // Set to true to automatically detect if libcamera is available. If true, UseLibcamera is ignored.
// }

func StartStreaming(ctx context.Context, dataChannel chan []byte) {
	log.Println("Start streaming")
	args := []string{
		"--inline", // H264: Force PPS/SPS header with every I frame
		"-t", "0",  // Disable timeout
		"-o", "-", // Output to stdout
		"--flush", // Flush output files immediately
		"--width", strconv.Itoa(640),
		"--height", strconv.Itoa(480),
		"--framerate", strconv.Itoa(60),
		"-n",                    // Do not show a preview window
		"--profile", "baseline", // H264 profile baseline, main or high
		//"--level", "4.2",
		//"--denoise", "cdn_off",
	}
	// if options.HorizontalFlip {
	// 	args = append(args, "--hflip")
	// }
	// if options.VerticalFlip {
	// 	args = append(args, "--vflip")
	// }
	// if options.Rotation != 0 {
	// 	args = append(args, "--rotation")
	// 	args = append(args, strconv.Itoa(options.Rotation))
	// }

	cmd := exec.CommandContext(ctx, "libcamera-vid", args...)
	defer cmd.Wait()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return
	}

	err = cmd.Start()
	if err != nil {
		log.Print(err)
		return
	}

	log.Println("Started libcamera-vid", cmd.Args)
	p := make([]byte, readBufferSize)
	buffer := make([]byte, bufferSizeKB*1024)
	currentPos := 0
	NALlen := len(nalSeparator)

	for {
		select {
		case <-ctx.Done():
			log.Println(ctx.Err())
			return
		default:
			n, err := stdout.Read(p)
			if err != nil {
				if err == io.EOF {
					log.Println("[libcamera-vid] EOF")
					return
				}
				log.Println(err)
			}

			copied := copy(buffer[currentPos:], p[:n])
			startPosSearch := currentPos - NALlen
			endPos := currentPos + copied

			if startPosSearch < 0 {
				startPosSearch = 0
			}
			nalIndex := bytes.Index(buffer[startPosSearch:endPos], nalSeparator)

			currentPos = endPos
			if nalIndex > 0 {
				nalIndex += startPosSearch

				// Boadcast before the NAL
				broadcast := make([]byte, nalIndex)
				copy(broadcast, buffer)
				dataChannel <- broadcast

				// Shift
				copy(buffer, buffer[nalIndex:currentPos])
				currentPos = currentPos - nalIndex
			}
		}
	}
}
