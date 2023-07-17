package carcam

import (
	"bytes"
	"context"
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
		"--width", c.options.width,
		"--height", c.options.height,
		"--framerate", c.options.fps,
		"-n", // Do not show a preview window
		//"--profile", c.options.profile, // H264 profile baseline, main or high
		//"--level", c.options.level,
	}
	// if c.options.horizontalFlip {
	// 	args = append(args, "--hflip")
	// }
	// if c.options.verticalFlip {
	// 	args = append(args, "--vflip")
	// }
	// if !c.options.deNoise {
	// 	args = append(args, "--denoise", "cdn_off")
	// }
	// if c.options.rotation != 0 {
	// 	args = append(args, "--rotation")
	// 	args = append(args, strconv.Itoa(c.options.rotation))
	// }

	cmd := exec.CommandContext(ctx, "libcamera-vid", args...)
	defer func() {
		log.Printf("killing cam streaming cmd...")
		if cmd.Process != nil {
			cmd.Process.Kill()
		} else {
			log.Printf("process was null")
		}
		cmd.Wait()
	}()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	log.Println("started libcamera-vid", cmd.Args)
	p := make([]byte, readBufferSize)
	buffer := make([]byte, bufferSizeKB*1024)
	currentPos := 0
	NALlen := len(nalSeparator)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			n, err := stdout.Read(p)
			if err != nil {
				// if err == io.EOF {
				// 	return fmt.Errorf("[libcamera-vid] EOF")
				// }
				return err
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
