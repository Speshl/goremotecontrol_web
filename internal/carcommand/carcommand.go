package carcommand

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

//pwm pins: 12,13,18,19,40,41,45

// esc/servo/pan/tilt
const escPinID = 12
const servoPinID = 13
const panPinID = 14
const tiltPinID = 15

const frequency = 64000
const maxvalue = uint32(255)
const midvalue = uint32(127)
const minvalue = uint32(0)

type CarCommand struct {
	name           string
	refreshRate    int
	CommandChannel chan []byte
	LatestCommand  LatestCommand
	pins           Pins
}

type Pins struct {
	escPin   rpio.Pin
	servoPin rpio.Pin
	panPin   rpio.Pin
	tiltPin  rpio.Pin
}

type LatestCommand struct {
	lock    sync.RWMutex
	used    bool
	command Command
}

type Command struct {
	esc   uint32
	servo uint32
	pan   uint32
	tilt  uint32
}

func NewCarCommand(name string, refreshRate int) *CarCommand {
	return &CarCommand{
		name:           name,
		refreshRate:    refreshRate,
		CommandChannel: make(chan []byte, 5),
		LatestCommand: LatestCommand{
			command: Command{
				esc:   midvalue,
				servo: midvalue,
				pan:   midvalue,
				tilt:  midvalue,
			},
		},
	}
}

func (c *CarCommand) Start(ctx context.Context) error {
	err := c.startGPIO()
	if err != nil {
		return err
	}
	defer rpio.Close()

	commandRate := 1000 / c.refreshRate
	commandDuration := time.Duration(int64(time.Millisecond) * int64(commandRate))
	commandTicker := time.NewTicker(commandDuration)
	seenSameCommand := 0
	warned := false
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("car command stopped: %s\n", ctx.Err())
		case command, ok := <-c.CommandChannel: //recieved command from client
			if !ok {
				return fmt.Errorf("car command channel stopped")
			}
			parsedCommand, err := c.parseCommand(command)
			if err != nil {
				log.Printf("WARNING: command failed to parse: %s\n", err)
				continue
			}
			c.LatestCommand.lock.Lock()
			c.LatestCommand.used = false
			c.LatestCommand.command = parsedCommand
			c.LatestCommand.lock.Unlock()
		case <-commandTicker.C: //time to send command to gpio
			c.LatestCommand.lock.RLock()
			if c.LatestCommand.used {
				c.LatestCommand.lock.RUnlock()
				if !warned {
					log.Printf("command was already used, skipping")
				}
				seenSameCommand++
				if seenSameCommand >= 5 {
					log.Println("no command, sending neutral")
					warned = true
					c.neutralCommand()
				}
				continue
			}
			seenSameCommand = 0
			warned = false
			c.LatestCommand.lock.RUnlock()

			c.LatestCommand.lock.Lock()
			c.LatestCommand.used = true
			command := c.LatestCommand.command
			c.LatestCommand.lock.Unlock()

			c.sendCommand(command)
		}
	}
}

func (c *CarCommand) startGPIO() error {
	err := rpio.Open()
	if err != nil {
		return err
	}

	c.pins.escPin = rpio.Pin(escPinID)
	c.pins.escPin.Mode(rpio.Pwm)
	c.pins.escPin.Freq(frequency)
	c.pins.escPin.DutyCycleWithPwmMode(midvalue, maxvalue, rpio.Balanced) //rpio.Markspace

	c.pins.servoPin = rpio.Pin(servoPinID)
	c.pins.servoPin.Mode(rpio.Pwm)
	c.pins.servoPin.Freq(frequency)
	c.pins.servoPin.DutyCycleWithPwmMode(midvalue, maxvalue, rpio.Balanced) //rpio.Markspace

	c.pins.tiltPin = rpio.Pin(tiltPinID)
	c.pins.tiltPin.Mode(rpio.Pwm)
	c.pins.tiltPin.Freq(frequency)
	c.pins.tiltPin.DutyCycleWithPwmMode(midvalue, maxvalue, rpio.Balanced) //rpio.Markspace

	c.pins.panPin = rpio.Pin(panPinID)
	c.pins.panPin.Mode(rpio.Pwm)
	c.pins.panPin.Freq(frequency)
	c.pins.panPin.DutyCycleWithPwmMode(midvalue, maxvalue, rpio.Balanced) //rpio.Markspace
	return nil
}

func (c *CarCommand) parseCommand(command []byte) (Command, error) {
	log.Printf("Command contained %d bytes", len(command))

	parsedCommand := Command{
		esc:   uint32(command[0]),
		servo: uint32(command[1]),
		pan:   uint32(command[2]),
		tilt:  uint32(command[3]),
	}

	log.Printf("Parsed Command: %+v", parsedCommand)
	return parsedCommand, nil
}

func (c *CarCommand) sendCommand(command Command) {
	c.pins.escPin.DutyCycleWithPwmMode(command.esc, maxvalue, rpio.Balanced)
	c.pins.escPin.DutyCycleWithPwmMode(command.servo, maxvalue, rpio.Balanced)
	c.pins.escPin.DutyCycleWithPwmMode(command.pan, maxvalue, rpio.Balanced)
	c.pins.escPin.DutyCycleWithPwmMode(command.tilt, maxvalue, rpio.Balanced)
}

func (c *CarCommand) neutralCommand() {
	c.pins.escPin.DutyCycleWithPwmMode(midvalue, maxvalue, rpio.Balanced)
	c.pins.escPin.DutyCycleWithPwmMode(midvalue, maxvalue, rpio.Balanced)
	c.pins.escPin.DutyCycleWithPwmMode(midvalue, maxvalue, rpio.Balanced)
	c.pins.escPin.DutyCycleWithPwmMode(midvalue, maxvalue, rpio.Balanced)
}
