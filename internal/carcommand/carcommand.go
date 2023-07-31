package carcommand

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

const frequency = 100000

const cycleLen = uint32(2000)
const maxvalue = uint32(250)
const midvalue = uint32(150)
const minvalue = uint32(50)

const escPinID = 12
const servoPinID = 13

// const panPinID = 14
// const tiltPinID = 15

type CarCommand struct {
	CommandChannel chan []byte
	options        CommandOptions
	latestCommand  LatestCommand
	pins           Pins
}

type CommandOptions struct {
	Name            string
	RefreshRate     int
	DisableCommands bool
	DeadZone        int

	MaxESC int
	MidESC int
	MinESC int

	MaxServo int
	MidServo int
	MinServo int
}

type Pins struct {
	esc   rpio.Pin
	servo rpio.Pin
	// pan   rpio.Pin
	// tilt  rpio.Pin
}

type LatestCommand struct {
	lock    sync.RWMutex
	used    bool
	command Command
}

type Command struct {
	esc   uint32
	servo uint32
	// pan   uint32
	// tilt  uint32
}

func NewCarCommand(options CommandOptions) *CarCommand {
	if options.DisableCommands {
		log.Println("Warning! GPIO commands are currently disabled")
	}
	return &CarCommand{
		CommandChannel: make(chan []byte, 5),
		options:        options,
		latestCommand: LatestCommand{
			command: Command{
				esc:   uint32(options.MidESC),
				servo: uint32(options.MidServo),
			},
		},
	}
}

func (c *CarCommand) Start(ctx context.Context) error {
	err := c.startGPIO()
	if err != nil {
		return err
	}
	defer func() {
		log.Println("closing rpio")
		err := rpio.Close()
		if err != nil {
			log.Printf("Error closing rpio: %s\n", err.Error())
		}
		log.Println("closed rpio")
	}()

	commandRate := 1000 / c.options.RefreshRate
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
			c.latestCommand.lock.Lock()
			c.latestCommand.used = false
			c.latestCommand.command = parsedCommand
			c.latestCommand.lock.Unlock()
		case <-commandTicker.C: //time to send command to gpio
			c.latestCommand.lock.RLock()
			if c.latestCommand.used {
				c.latestCommand.lock.RUnlock()
				if !warned {
					//log.Printf("command was already used, skipping")
				}
				seenSameCommand++
				if seenSameCommand >= 1000 {
					if !warned {
						log.Println("no command, sending neutral")
					}
					warned = true
					c.sendNeutral()
				}
			} else {
				seenSameCommand = 0
				warned = false
				c.latestCommand.lock.RUnlock()

				c.latestCommand.lock.Lock()
				c.latestCommand.used = true
				command := c.latestCommand.command
				c.latestCommand.lock.Unlock()

				c.sendCommand(command)
			}
		}
	}
}

func (c *CarCommand) startGPIO() error {
	err := rpio.Open()
	if err != nil {
		return err
	}

	c.pins.esc = rpio.Pin(escPinID)
	c.pins.esc.Mode(rpio.Pwm)
	c.pins.esc.Freq(frequency)

	c.pins.servo = rpio.Pin(servoPinID)
	c.pins.servo.Mode(rpio.Pwm)
	c.pins.servo.Freq(frequency)
	c.sendNeutral()
	return nil
}

func (c *CarCommand) parseCommand(command []byte) (Command, error) {
	parsedCommand := Command{
		esc:   c.mapToRange(uint32(command[0]), 0, 255, uint32(c.options.MinESC), uint32(c.options.MaxESC)),
		servo: c.mapToRange(uint32(command[1]), 0, 255, uint32(c.options.MinServo), uint32(c.options.MaxServo)),
	}

	parsedCommand = c.applyDeadZone(parsedCommand)
	return parsedCommand, nil
}

func (c *CarCommand) sendNeutral() {
	if !c.options.DisableCommands {
		c.pins.esc.DutyCycle(uint32(c.options.MidESC), cycleLen)
		c.pins.servo.DutyCycle(uint32(c.options.MidServo), cycleLen)
		// c.pins.pan.DutyCycle(command.pan, cycleLen)
		// c.pins.tilt.DutyCycle(command.tilt, cycleLen)
	}
}

func (c *CarCommand) sendCommand(command Command) {
	if !c.options.DisableCommands {
		//log.Printf("Sending Command: %+v", command)
		c.pins.esc.DutyCycle(command.esc, cycleLen)
		c.pins.servo.DutyCycle(command.servo, cycleLen)
		// c.pins.pan.DutyCycle(command.pan, cycleLen)
		// c.pins.tilt.DutyCycle(command.tilt, cycleLen)
	}
}

func (c *CarCommand) mapToRange(value, min, max, minReturn, maxReturn uint32) uint32 {
	return (maxReturn-minReturn)*(value-min)/(max-min) + minReturn
}

func (c *CarCommand) applyDeadZone(command Command) Command {
	returnCommand := command

	if command.esc > midvalue && midvalue+uint32(c.options.DeadZone) > command.esc {
		returnCommand.esc = midvalue
	}

	if command.esc < midvalue && midvalue-uint32(c.options.DeadZone) < command.esc {
		returnCommand.esc = midvalue
	}

	if command.servo > midvalue && midvalue+uint32(c.options.DeadZone) > command.servo {
		returnCommand.servo = midvalue
	}

	if command.servo < midvalue && midvalue-uint32(c.options.DeadZone) < command.servo {
		returnCommand.servo = midvalue
	}

	// if command.pan > midvalue && midvalue+deadZone > command.pan {
	// 	returnCommand.pan = midvalue
	// }

	// if command.pan < midvalue && midvalue-deadZone < command.pan {
	// 	returnCommand.pan = midvalue
	// }

	// if command.tilt > midvalue && midvalue+deadZone > command.tilt {
	// 	returnCommand.tilt = midvalue
	// }

	// if command.tilt < midvalue && midvalue-deadZone < command.tilt {
	// 	returnCommand.tilt = midvalue
	// }

	return returnCommand
}
