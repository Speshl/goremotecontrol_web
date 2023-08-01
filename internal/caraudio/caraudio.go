package caraudio

import (
	"context"
	"io"
	"os/exec"
)

type readerAtSeeker interface {
	io.Reader
	io.ReaderAt
	io.Seeker
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

func (c *CarAudio) Play(ctx context.Context) error {
	args := []string{
		"~/scripts/starwars.wav",
	}
	cmd := exec.CommandContext(ctx, "aplay", args...)
	err := cmd.Start()
	if err != nil {
		return err
	}
	defer cmd.Wait()
	return nil
}
