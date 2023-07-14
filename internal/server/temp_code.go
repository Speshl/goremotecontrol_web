package server

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"log"
	"os"
	"time"

	x264 "github.com/gen2brain/x264-go"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"
	"github.com/pion/webrtc/v3/pkg/media/oggreader"
	"gocv.io/x/gocv"
)

var opts = &x264.Options{
	Width:     640,
	Height:    480,
	FrameRate: 25,
	Tune:      "zerolatency",
	Preset:    "ultrafast",
	Profile:   "main",
	LogLevel:  x264.LogDebug,
}

func TempRecordPainted() (err error) {
	file, err := os.Create("screen.264")
	if err != nil {
		return fmt.Errorf("Error creating video file: %w", err)
	}

	enc, err := x264.NewEncoder(file, opts)
	if err != nil {
		return fmt.Errorf("Error creating encoder: %w", err)
	}

	frameTicker := time.NewTicker(time.Second / time.Duration(25))

	defer func() {
		err = enc.Flush()
		if err != nil {
			err = fmt.Errorf("Error flushing encoder: %w", err)
			return
		}

		file.Close()
		if err != nil {
			err = fmt.Errorf("Error flushing encoder: %w", err)
		}

	}()

	frameCounter := 0
	for range frameTicker.C {
		if frameCounter > 500 {
			frameTicker.Stop()
			return nil
		}
		img := x264.NewYCbCr(image.Rect(0, 0, opts.Width, opts.Height))
		draw.Draw(img, img.Bounds(), image.Black, image.ZP, draw.Src)
		img.Set(frameCounter, opts.Height/2, color.RGBA{255, 0, 0, 255})

		log.Println("Encoding frame %d\n", frameCounter)
		frameCounter++
		err = enc.Encode(img)
		if err != nil {
			return fmt.Errorf("Error encoding frame: %w", err)
		}
	}
	return nil
}

func PlayReadWebCam() error {
	log.Printf("Start Reading Webcam")
	defer log.Printf("Done Reading Webcam")

	deviceID := 0
	webcam, err := gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		log.Printf("Failing opening video capture device: %s\n", err.Error())
		return err
	}

	img := gocv.NewMat()
	ok := webcam.Read(&img)
	if !ok {
		return fmt.Errorf("error reading from video devide %d\n", deviceID)
	}

	size := img.Size()
	log.Printf("Frame Empty: %t Frame Size - 0: %d 1: %d\n", img.Empty(), size[0], size[1])
	return nil
}

func (c *Connection) PlayTempAudio(ctx context.Context) error {
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

func (c *Connection) TempStreamVideo(ctx context.Context) error {
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
			nal, h264Err := fileReader.NextNAL()
			if h264Err == io.EOF {
				fmt.Printf("All video frames parsed and sent")
				return
			}
			if h264Err != nil {
				log.Printf("error reading next nal: %s\n", err.Error())
				return
			}

			//log.Printf("Sending frame - NAL - %+v\n", nal.Data)
			if h264Err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: frameDuration}); h264Err != nil {
				log.Printf("error writing sample: %s\n", err.Error())
				return
			}
		}
	}()
	return nil
}

func TempRecordCam() (err error) {
	deviceID := 0
	webcam, err := gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		log.Printf("Failing opening video capture device: %s\n", err.Error())
		return err
	}

	file, err := os.Create("screen.264")
	if err != nil {
		return fmt.Errorf("Error creating video file: %w", err)
	}

	enc, err := x264.NewEncoder(file, opts)
	if err != nil {
		return fmt.Errorf("Error creating encoder: %w", err)
	}

	frameTicker := time.NewTicker(time.Second / time.Duration(25))

	defer func() {
		err = enc.Flush()
		if err != nil {
			err = fmt.Errorf("Error flushing encoder: %w", err)
			return
		}

		file.Close()
		if err != nil {
			err = fmt.Errorf("Error flushing encoder: %w", err)
		}

	}()

	frameCounter := 0
	for range frameTicker.C {
		if frameCounter > 500 {
			frameTicker.Stop()
			return nil
		}

		matImg := gocv.NewMat()
		ok := webcam.Read(&matImg)
		if !ok {
			return fmt.Errorf("error reading from video devide %d\n", deviceID)
		}

		img, err := matImg.ToImage()
		if err != nil {
			return fmt.Errorf("error converting from gocv mat to image: %w", err)
		}

		// img := x264.NewYCbCr(image.Rect(0, 0, opts.Width, opts.Height))
		// draw.Draw(img, img.Bounds(), image.Black, image.ZP, draw.Src)
		// img.Set(frameCounter, opts.Height/2, color.RGBA{255, 0, 0, 255})

		log.Println("Encoding frame %d\n", frameCounter)
		frameCounter++
		err = enc.Encode(img)
		if err != nil {
			return fmt.Errorf("Error encoding frame: %w", err)
		}
	}
	return nil
}
