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

const deadZone = uint32(2)

const frequency = 100000
const cycleLen = uint32(2000)

const maxvalue_limited = uint32(200)
const maxvalue = uint32(250)

const midvalue = uint32(150)

const minvalue_limited = uint32(100)
const minvalue = uint32(50)

type CarCommand struct {
	name            string
	refreshRate     int
	disableCommands bool
	CommandChannel  chan []byte
	LatestCommand   LatestCommand
	pins            Pins
}

type Pins struct {
	esc   rpio.Pin
	servo rpio.Pin
	pan   rpio.Pin
	tilt  rpio.Pin
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

func NewCarCommand(name string, refreshRate int, disableCommands bool) *CarCommand {
	if disableCommands {
		log.Println("Warning! GPIO commands are currently disabled")
	}
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
	defer func() {
		log.Println("closing rpio")
		err := rpio.Close()
		if err != nil {
			log.Printf("Error closing rpio: %s\n", err.Error())
		}
		log.Println("closed rpio")
	}()

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
					//log.Printf("command was already used, skipping")
				}
				seenSameCommand++
				if seenSameCommand >= 5 {
					if !warned {
						log.Println("no command, sending neutral")
					}
					warned = true
					c.sendNeutral()
				}
			} else {
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
}

func (c *CarCommand) startGPIO() error {
	err := rpio.Open()
	if err != nil {
		return err
	}

	c.pins.esc = rpio.Pin(escPinID)
	//c.pins.esc.Pwm()
	c.pins.esc.Mode(rpio.Pwm)
	c.pins.esc.Freq(frequency)

	c.pins.servo = rpio.Pin(servoPinID)
	// c.pins.servo.Pwm()
	c.pins.servo.Mode(rpio.Pwm)
	c.pins.servo.Freq(frequency)

	// c.pins.tilt = rpio.Pin(tiltPinID)
	// // c.pins.tilt.Pwm()
	// c.pins.tilt.Mode(rpio.Pwm)
	// c.pins.tilt.Freq(frequency)

	// c.pins.pan = rpio.Pin(panPinID)
	// // c.pins.pan.Pwm()
	// c.pins.pan.Mode(rpio.Pwm)
	// c.pins.pan.Freq(frequency)
	c.sendNeutral()
	return nil
}

func (c *CarCommand) parseCommand(command []byte) (Command, error) {
	parsedCommand := Command{
		esc:   c.mapToRange(uint32(command[0]), 0, 255, minvalue_limited, maxvalue_limited),
		servo: c.mapToRange(uint32(command[1]), 0, 255, minvalue_limited, maxvalue_limited),
		pan:   c.mapToRange(uint32(command[2]), 0, 255, minvalue_limited, maxvalue_limited),
		tilt:  c.mapToRange(uint32(command[3]), 0, 255, minvalue_limited, maxvalue_limited),
	}

	parsedCommand = c.applyDeadZone(parsedCommand)

	if parsedCommand.esc != midvalue || parsedCommand.servo != midvalue ||
		parsedCommand.pan != midvalue || parsedCommand.tilt != midvalue {
		log.Printf("Parsed Command: %+v", parsedCommand)
	}
	return parsedCommand, nil
}

func (c *CarCommand) sendNeutral() {
	c.sendCommand(Command{
		esc:   midvalue,
		servo: midvalue,
		tilt:  midvalue,
		pan:   midvalue,
	})
}

func (c *CarCommand) sendCommand(command Command) {
	if !c.disableCommands {
		log.Printf("Sending Command: %+v", command)
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

	if command.esc > midvalue && midvalue+deadZone > command.esc {
		returnCommand.esc = midvalue
	}

	if command.esc < midvalue && midvalue-deadZone < command.esc {
		returnCommand.esc = midvalue
	}

	if command.servo > midvalue && midvalue+deadZone > command.servo {
		returnCommand.servo = midvalue
	}

	if command.servo < midvalue && midvalue-deadZone < command.servo {
		returnCommand.servo = midvalue
	}

	if command.pan > midvalue && midvalue+deadZone > command.pan {
		returnCommand.pan = midvalue
	}

	if command.pan < midvalue && midvalue-deadZone < command.pan {
		returnCommand.pan = midvalue
	}

	if command.tilt > midvalue && midvalue+deadZone > command.tilt {
		returnCommand.tilt = midvalue
	}

	if command.tilt < midvalue && midvalue-deadZone < command.tilt {
		returnCommand.tilt = midvalue
	}

	return returnCommand
}
