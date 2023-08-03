package carspeaker

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"time"
)

const delayBetweenSounds = 2 * time.Second

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
	SpeakerChannel chan string
	lastPlayedAt   time.Time
	options        SpeakerOptions
}

type SpeakerOptions struct {
	Name string
}

func NewCarSpeaker(options SpeakerOptions) (*CarSpeaker, error) {
	return &CarSpeaker{
		SpeakerChannel: make(chan string, 10),
		options:        options,
	}, nil
}

func (c *CarSpeaker) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("speaker listener done due to ctx")
			return nil
		case data, ok := <-c.SpeakerChannel:
			log.Printf("Got sound %s\n", data)
			if !ok {
				log.Println("speaker listener channel closed, stopping")
				return nil
			}
			if c.lastPlayedAt.Add(delayBetweenSounds).Compare(time.Now()) == 1 {
				continue //Skip playing sound so we don't spam
			}

			c.lastPlayedAt = time.Now()
			go func() {
				err := c.PlayFromGroup(ctx, data)
				if err != nil {
					log.Printf("failed to play sound from group - %s\n", err.Error())
				}
			}()
		}
	}
}

func (c *CarSpeaker) PlayFromGroup(ctx context.Context, group string) error {
	soundGroup, ok := soundGroups[group]
	if !ok {
		return fmt.Errorf("error: sound group not found")
	}

	value := rand.Intn(len(soundGroup))
	return c.Play(ctx, soundGroup[value])
}

func (c *CarSpeaker) Play(ctx context.Context, sound string) error {

	soundPath, ok := soundMap[sound]
	if !ok {
		return fmt.Errorf("error: sound not found")
	}

	log.Printf("start playing %s sound\n", sound)
	defer log.Printf("finished playing %s sound\n", sound)
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
