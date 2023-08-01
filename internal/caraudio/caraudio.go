package caraudio

import (
	"context"
	"fmt"
	"log"
	"os/exec"
)

type CarAudio struct {
	options AudioOptions
}

type AudioOptions struct {
	Name string
}

func NewCarAudio(options AudioOptions) (*CarAudio, error) {
	return &CarAudio{
		options: options,
	}, nil
}

func (c *CarAudio) Play(ctx context.Context) error {
	log.Println("Start playing Star Wars")
	args := []string{
		// "-c",
		"./play.sh",
		"./internal/caraudio/starwars.wav",
	}
	cmd := exec.Command("/bin/sh", args...)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting audio playback - %w", err)
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error during audio playback - %w", err)
	}
	return nil
}
