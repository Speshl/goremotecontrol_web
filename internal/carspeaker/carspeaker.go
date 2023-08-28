package carspeaker

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"sync"
	"time"
)

const DelayBetweenSounds = 2 * time.Second

const DefaultDevice = "0"
const DefaultVolume = "5.0"

var soundMap = map[string]string{
	//Affirmatives
	"kool_aid_oh_yeah":  "./internal/carspeaker/audio/kool_aid_oh_yeah.wav",
	"hell_yeah_brother": "./internal/carspeaker/audio/hell_yeah_brother.wav",
	"yeah":              "./internal/carspeaker/audio/yeah.wav",

	//Negatives
	"negative_ghostrider": "./internal/carspeaker/audio/negative_ghostrider.wav",
	"oh_hell_no":          "./internal/carspeaker/audio/oh_hell_no.wav",

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
	"negative":    {"oh_hell_no", "negative_ghostrider"},
	"aggressive":  {"move_bitch", "emotional_damage", "bruh", "spongebob_fail"},
	"sorry":       {""},
}

type CarSpeaker struct {
	SpeakerChannel chan string
	config         SpeakerConfig
	lock           sync.RWMutex
	playing        bool
}

type SpeakerConfig struct {
	Device string
	Volume string
}

func NewCarSpeaker(cfg SpeakerConfig) (*CarSpeaker, error) {
	carSpeaker := CarSpeaker{
		SpeakerChannel: make(chan string, 10),
		config:         cfg,
	}
	return &carSpeaker, nil
}

func (c *CarSpeaker) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("speaker listener done due to ctx")
			return nil
		case data, ok := <-c.SpeakerChannel:
			if !ok {
				log.Println("speaker listener channel closed, stopping")
				return nil
			}

			gotLock := c.lock.TryLock()
			if !gotLock {
				continue
			}

			go func() {
				defer c.lock.Unlock()
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

	log.Printf("start playing %s sound\n", sound)
	defer log.Printf("finished playing %s sound\n", sound)

	soundPath, ok := soundMap[sound]
	if !ok {
		return fmt.Errorf("error: sound not found")
	}

	args := []string{
		"-D", "hw:CARD=wm8960soundcard,DEV=0", //TODO: Make these changeable by environment variable
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
