package carcam

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
)

const readBufferSize = 4096
const bufferSizeKB = 256

var nalSeparator = []byte{0, 0, 0, 1} //NAL break

func (c *CarCam) StartStreaming(ctx context.Context) error {
	log.Println("start streaming...")
	args := []string{
		"--inline", // H264: Force PPS/SPS header with every I frame
		"-t", "0",  // Disable timeout
		"-o", "-", // Output to stdout
		"--flush", // Flush output files immediately
		"--width", c.config.Width,
		"--height", c.config.Height,
		"--framerate", c.config.Fps,
		"-n",                          // Do not show a preview window
		"--profile", c.config.Profile, // H264 profile baseline, main or high
		//"--level", c.config.level,
	}
	if c.config.HorizontalFlip {
		args = append(args, "--hflip")
	}
	if c.config.VerticalFlip {
		args = append(args, "--vflip")
	}

	if c.config.Mode != "" {
		args = append(args, "--mode", c.config.Mode)
	}
	// if !c.config.deNoise {
	// 	args = append(args, "--denoise", "cdn_off")
	// }
	// if c.config.rotation != 0 {
	// 	args = append(args, "--rotation")
	// 	args = append(args, strconv.Itoa(c.config.rotation))
	// }

	if c.config.DisableVideo {
		return c.noVideoLoop(ctx)
	}

	cmd := exec.CommandContext(ctx, "libcamera-vid", args...)
	defer func() {
		log.Println("killing cam streaming cmd...")
		if cmd.Process != nil {
			err := cmd.Process.Kill()
			if err != nil {
				log.Printf("Error killing cam process: %s", err.Error())
			}
		} else {
			log.Println("process was null")
		}
		cmd.Wait()
		log.Println("killed cam streaming cmd")
	}()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed getting std out pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed starting camera: %w", err)
	}

	log.Println("started libcamera-vid", cmd.Args)
	p := make([]byte, readBufferSize)
	buffer := make([]byte, bufferSizeKB*1024)
	currentPos := 0
	NALlen := len(nalSeparator)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping cam due to context")
			return ctx.Err()
		default:
			n, err := stdout.Read(p)
			if err != nil {
				// if err == io.EOF {
				// 	return fmt.Errorf("[libcamera-vid] EOF")
				// }
				return fmt.Errorf("failed reading camera from std out: %w", err)
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
				c.videoChannel <- broadcast

				// Shift
				copy(buffer, buffer[nalIndex:currentPos])
				currentPos = currentPos - nalIndex
			}
		}
	}
}

func (c *CarCam) noVideoLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping cam due to context")
			return ctx.Err()
		}
	}
}
