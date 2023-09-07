package carspeaker

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"

	"github.com/Speshl/goremotecontrol_web/internal/gst"
	"github.com/pion/webrtc/v3"
)

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
	"startup":             "./internal/carspeaker/audio/startup.wav",
	"shutdown":            "./internal/carspeaker/audio/shutting_down.wav",
	"client_connected":    "./internal/carspeaker/audio/jarvis_connected.wav",
	"client_disconnected": "./internal/carspeaker/audio/jarvis_disconnected.wav",
}

var soundGroups = map[string][]string{
	"affirmative": {"kool_aid_oh_yeah", "hell_yeah_brother", "yeah"},
	"negative":    {"oh_hell_no", "negative_ghostrider"},
	"aggressive":  {"move_bitch", "emotional_damage", "bruh", "spongebob_fail"},
	"sorry":       {""},
}

type CarSpeaker struct {
	MemeSoundChannel chan string
	//ClientAudioChannel chan []byte
	config  SpeakerConfig
	lock    sync.RWMutex
	playing bool
}

type SpeakerConfig struct {
	Device string
	Volume string
}

func NewCarSpeaker(cfg SpeakerConfig) (*CarSpeaker, error) {
	carSpeaker := CarSpeaker{
		MemeSoundChannel: make(chan string, 10),
		//ClientAudioChannel: make(chan []byte, 10),
		config: cfg,
	}
	return &carSpeaker, nil
}

func (c *CarSpeaker) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("speaker listener done due to ctx")
			return nil
		case data, ok := <-c.MemeSoundChannel:
			if !ok {
				log.Println("speaker listener channel closed, stopping")
				return nil
			}

			go func() {
				err := c.PlaySound(ctx, data)
				if err != nil {
					log.Printf("failed to play sound - %s\n", err.Error())
				}
			}()
		}
	}
}

func (c *CarSpeaker) TrackPlayer(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	log.Println("start playing client track")
	defer log.Println("done playing client track")
	// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval. Not sure what this means or if I need it?
	// go func() {
	// 	ticker := time.NewTicker(time.Second * 3)
	// 	for range ticker.C {
	// 		if err := c.CTX.Err(); err != nil {
	// 			return
	// 		}
	// 		rtcpSendErr := c.PeerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
	// 		if rtcpSendErr != nil {
	// 			log.Printf("error sending keyframe on ticker - %w\n", rtcpSendErr)
	// 			return
	// 		}
	// 	}
	// }()
	codecName := strings.Split(track.Codec().RTPCodecCapability.MimeType, "/")[1]
	log.Printf("Track has started, of type %d: %s \n", track.PayloadType(), codecName)
	pipeline := gst.CreateRecievePipeline(track.PayloadType(), strings.ToLower(codecName), c.config.Device, c.config.Volume)
	pipeline.Start()
	defer pipeline.Stop()

	buf := make([]byte, 1400)
	for {
		i, _, err := track.Read(buf)
		if err != nil {
			log.Printf("stopping client audio - error reading client audio track buffer - %s\n", err)
			return
		}
		//log.Printf("Pushing %d bytes to pipeline", i)
		pipeline.Push(buf[:i])
	}
}

// Plays a named sound, if the sound is a group name, will randomly play a sound from that group
func (c *CarSpeaker) PlaySound(ctx context.Context, sound string) error {
	_, ok := soundGroups[sound]
	if !ok {
		return c.Play(ctx, sound)
	}
	return c.PlayFromGroup(ctx, sound)

}

// Plays a random prebaked meme sound from the specified group (affirmative, negative, aggressive, sorry)
func (c *CarSpeaker) PlayFromGroup(ctx context.Context, group string) error {
	soundGroup, ok := soundGroups[group]
	if !ok {
		return fmt.Errorf("error: sound group not found")
	}

	value := rand.Intn(len(soundGroup))
	return c.Play(ctx, soundGroup[value])
}

// Plays a named sound from the soundMap
func (c *CarSpeaker) Play(ctx context.Context, sound string) error {

	gotLock := c.lock.TryLock()
	if !gotLock {
		if sound == "client_disconnected" {
			c.lock.Lock() //Wait for the lock to end if its the shutdown signal, should be unlocked soon
		} else {
			return fmt.Errorf("speaker is locked, other sound is playing")
		}
	}
	defer c.lock.Unlock()

	log.Printf("start playing %s sound\n", sound)
	defer log.Printf("finished playing %s sound\n", sound)

	// soundPath, ok := soundMap[sound]
	// if !ok {
	// 	return fmt.Errorf("error: sound not found")
	// }

	// args := []string{
	// 	//"-D", "hw:CARD=wm8960soundcard,DEV=0", //TODO: Make these changeable by environment variable
	// 	"-E",
	// 	"aplay",
	// 	"-D", "hw:CARD=wm8960soundcard,DEV=0",
	// 	soundPath,
	// }
	// cmd := exec.CommandContext(ctx, "sudo", args...)
	// err := cmd.Start()
	// if err != nil {
	// 	return fmt.Errorf("error starting audio playback - %w", err)
	// }
	// err = cmd.Wait()
	// if err != nil {
	// 	return fmt.Errorf("error during audio playback - %w", err)
	// }
	return nil
}
