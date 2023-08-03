package carspeaker

import (
	"context"
	"fmt"
	"log"
	"os/exec"
)

var soundMap = map[string]string{
	//Affirmatives
	"kool_aid_oh_yeah":  "./internal/carspeaker/audio/kool_aid_oh_yeah.wav",
	"hell_yeah_brother": "./internal/carspeaker/audio/hell_yeah_brother.wav",
	"yeah":              "./internal/carspeaker/audio/yeah.wav",

	//Negatives
	"oh_hell_no": "./internal/carspeaker/audio/oh_hell_no.wav",
	"nope":       "./internal/carspeaker/audio/nope.wav",

	//Aggressive
	"move_bitch":       "./internal/carspeaker/audio/move_bitch.wav",
	"emotional_damage": "./internal/carspeaker/audio/emotional_damage.wav",
	"bruh":             "./internal/carspeaker/audio/bruh.wav",
	"spongebob_fail":   "./internal/carspeaker/audio/spongebob_fail.wav",

	//Sorry

	//other
	"startup": "./internal/carspeaker/audio/startup.wav",
}

var soundGroups = map[string][]string{
	"affirmative": {"kool_aid_oh_yeah", "hell_yeah_brother", "yeah"},
	"negative":    {"oh_hell_no", "nope"},
	"aggressive":  {"move_bitch", "emotional_damage", "bruh", "spongebob_fail"},
	"sorry":       {""},
}

type CarSpeaker struct {
	options SpeakerOptions
}

type SpeakerOptions struct {
	Name string
}

func NewCarSpeaker(options SpeakerOptions) (*CarSpeaker, error) {
	return &CarSpeaker{
		options: options,
	}, nil
}

func (c *CarSpeaker) Play(ctx context.Context, sound string) error {

	soundPath, ok := soundMap[sound]
	if !ok {
		return fmt.Errorf("error: sound not found")
	}

	log.Printf("start playing %s sound\n", sound)
	defer log.Printf("finished playing %s sound\n", sound)
	// args := []string{
	// 	"./play.sh",
	// 	soundPath,
	// }
	// cmd := exec.CommandContext(ctx, "/bin/sh", args...)
	args := []string{
		"-D", "hw:CARD=wm8960soundcard,DEV=0",
		soundPath,
	}
	cmd := exec.CommandContext(ctx, "aplay", args...)
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
