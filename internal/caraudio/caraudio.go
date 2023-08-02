package caraudio

import (
	"context"
	"fmt"
	"log"
	"os/exec"
)

var soundMap = map[string]string{
	"startup": "./internal/caraudio/audio/startup.wav",
}

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

func (c *CarAudio) Play(ctx context.Context, sound string) error {

	soundPath, ok := soundMap[sound]
	if !ok {
		return fmt.Errorf("sound not found")
	}

	log.Printf("start playing %s\n", sound)
	args := []string{
		"./play.sh",
		soundPath,
	}
	cmd := exec.CommandContext(ctx, "/bin/sh", args...)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting audio playback - %w", err)
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error during audio playback - %w", err)
	}
	log.Printf("finished playing %s\n", sound)
	return nil
}
