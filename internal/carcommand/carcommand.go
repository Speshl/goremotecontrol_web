package carcommand

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/googolgl/go-i2c"
	"github.com/googolgl/go-pca9685"
)

const MaxValue = 255
const MinValue = 0
const MidValue = (MaxValue - MinValue) / 2

type CarCommand struct {
	CommandChannel chan []byte
	options        CommandOptions
	latestCommand  LatestCommand

	servoController *pca9685.PCA9685
	servos          Servos
}

type Servos struct {
	esc   *pca9685.Servo
	steer *pca9685.Servo
	pan   *pca9685.Servo
	tilt  *pca9685.Servo
}

// ServoRangeDef    int     = 135
// ServoMinPulseDef float32 = 750.0
// ServoMaxPulseDef float32 = 2250.0

type CommandOptions struct {
	RefreshRate     int
	DisableCommands bool
	DeadZone        int

	ESCChannel  int
	MaxESCPulse float32
	MinESCPulse float32
	ESCLimit    uint32 //Subtracted from max, and added to min to keep off servo endpoints

	SteerChannel  int
	MaxSteerPulse float32
	MinSteerPulse float32
	SteerLimit    uint32 //Subtracted from max, and added to min to keep off servo endpoints

	PanChannel  int
	MaxPanPulse float32
	MinPanPulse float32
	PanLimit    uint32 //Subtracted from max, and added to min to keep off servo endpoints

	TiltChannel  int
	MaxTiltPulse float32
	MinTiltPulse float32
	TiltLimit    uint32 //Subtracted from max, and added to min to keep off servo endpoints
}

type LatestCommand struct {
	// lock    sync.RWMutex
	used    bool
	command Command
}

type Command struct {
	esc   uint32
	steer uint32
	pan   uint32
	tilt  uint32
}

func NewCarCommand(options *CommandOptions) (*CarCommand, error) {
	if options.DisableCommands {
		log.Println("Warning! GPIO commands are currently disabled")
	}
	carCommand := CarCommand{
		CommandChannel: make(chan []byte, 5),
		options: CommandOptions{
			RefreshRate:     60,
			DisableCommands: false,
			DeadZone:        2,

			ESCChannel:  3,
			ESCLimit:    50,
			MaxESCPulse: pca9685.ServoMaxPulseDef,
			MinESCPulse: pca9685.ServoMinPulseDef,

			SteerChannel:  4,
			SteerLimit:    50,
			MaxSteerPulse: pca9685.ServoMaxPulseDef,
			MinSteerPulse: pca9685.ServoMinPulseDef,

			PanChannel:  2,
			PanLimit:    0,
			MaxPanPulse: pca9685.ServoMaxPulseDef,
			MinPanPulse: pca9685.ServoMinPulseDef,

			TiltChannel:  1,
			TiltLimit:    0,
			MaxTiltPulse: pca9685.ServoMaxPulseDef,
			MinTiltPulse: pca9685.ServoMinPulseDef,
		},
		latestCommand: LatestCommand{
			command: Command{
				esc:   MidValue,
				steer: MidValue,
				pan:   MidValue,
				tilt:  MidValue,
			},
		},
	}

	if options != nil {
		carCommand.options = *options
	}

	err := carCommand.SetupServoController()
	if err != nil {
		return nil, fmt.Errorf("failed setting up servo controller - %w", err)
	}

	return &carCommand, nil
}

func (c *CarCommand) SetupServoController() error {
	i2c, err := i2c.New(pca9685.Address, "/dev/i2c-1")
	if err != nil {
		return fmt.Errorf("error starting i2c with address - %w", err)
	}

	c.servoController, err = pca9685.New(i2c, nil)
	if err != nil {
		return fmt.Errorf("error getting servo driver - %w", err)
	}

	c.servos.esc = c.servoController.ServoNew(c.options.ESCChannel, &pca9685.ServOptions{
		AcRange:  pca9685.ServoRangeDef,
		MinPulse: c.options.MinESCPulse,
		MaxPulse: c.options.MaxESCPulse,
	})

	c.servos.steer = c.servoController.ServoNew(c.options.SteerChannel, &pca9685.ServOptions{
		AcRange:  pca9685.ServoRangeDef,
		MinPulse: c.options.MinSteerPulse,
		MaxPulse: c.options.MaxSteerPulse,
	})

	c.servos.pan = c.servoController.ServoNew(c.options.PanChannel, &pca9685.ServOptions{
		AcRange:  pca9685.ServoRangeDef,
		MinPulse: c.options.MinPanPulse,
		MaxPulse: c.options.MaxPanPulse,
	})

	c.servos.tilt = c.servoController.ServoNew(c.options.TiltChannel, &pca9685.ServOptions{
		AcRange:  pca9685.ServoRangeDef,
		MinPulse: c.options.MinTiltPulse,
		MaxPulse: c.options.MaxTiltPulse,
	})

	return nil
}

func (c *CarCommand) Start(ctx context.Context) error {
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
			c.latestCommand.used = false
			c.latestCommand.command = parsedCommand
		case <-commandTicker.C: //time to send command to gpio

			if c.latestCommand.used {
				seenSameCommand++
				if seenSameCommand >= 1000 {
					if !warned {
						log.Println("no command, sending neutral")
					}
					warned = true
					err := c.sendNeutral()
					if err != nil {
						return fmt.Errorf("error sending neutral command: %w", err)
					}
				}
			} else {
				seenSameCommand = 0
				warned = false
				c.latestCommand.used = true
				command := c.latestCommand.command
				err := c.sendCommand(command)
				if err != nil {
					return fmt.Errorf("error sending neutral command: %w", err)
				}
			}
		}
	}
}

func (c *CarCommand) parseCommand(command []byte) (Command, error) {
	if len(command) < 4 {
		return Command{}, fmt.Errorf("incorrect command length - %d", len(command))
	}
	parsedCommand := Command{
		esc:   c.mapToRange(uint32(command[0]), MinValue, MaxValue, MinValue+c.options.ESCLimit, MaxValue-c.options.ESCLimit),
		steer: c.mapToRange(uint32(command[1]), MinValue, MaxValue, MinValue+c.options.SteerLimit, MaxValue-c.options.SteerLimit),
		pan:   c.mapToRange(uint32(command[2]), MinValue, MaxValue, MinValue+c.options.PanLimit, MaxValue-c.options.PanLimit),
		tilt:  c.mapToRange(uint32(command[3]), MinValue, MaxValue, MinValue+c.options.TiltLimit, MaxValue-c.options.TiltLimit),
	}

	return c.applyDeadZone(parsedCommand), nil
}

func (c *CarCommand) sendNeutral() error {
	if !c.options.DisableCommands {
		err := c.servos.esc.Fraction(0.5)
		if err != nil {
			log.Printf("failed sending esc command: %s\n", err.Error())
			return err
		}

		err = c.servos.steer.Fraction(0.5)
		if err != nil {
			log.Printf("failed sending steer command: %s\n", err.Error())
			return err
		}

		err = c.servos.pan.Fraction(0.5)
		if err != nil {
			log.Printf("failed sending pan command: %s\n", err.Error())
			return err
		}

		err = c.servos.tilt.Fraction(0.5)
		if err != nil {
			log.Printf("failed sending tilt command: %s\n", err.Error())
			return err
		}
	}
}

func (c *CarCommand) sendCommand(command Command) error {
	if !c.options.DisableCommands {
		err := c.servos.esc.Fraction(float32(command.esc) / MaxValue)
		if err != nil {
			log.Printf("failed sending esc command: %s\n", err.Error())
			return err
		}

		err = c.servos.steer.Fraction(float32(command.steer) / MaxValue)
		if err != nil {
			log.Printf("failed sending steer command: %s\n", err.Error())
			return err
		}

		err = c.servos.pan.Fraction(float32(command.pan) / MaxValue)
		if err != nil {
			log.Printf("failed sending pan command: %s\n", err.Error())
			return err
		}

		err = c.servos.tilt.Fraction(float32(command.tilt) / MaxValue)
		if err != nil {
			log.Printf("failed sending tilt command: %s\n", err.Error())
			return err
		}
	}
	return nil
}

func (c *CarCommand) mapToRange(value, min, max, minReturn, maxReturn uint32) uint32 {
	return (maxReturn-minReturn)*(value-min)/(max-min) + minReturn
}

func (c *CarCommand) applyDeadZone(command Command) Command {
	returnCommand := command

	if command.esc > MidValue && MidValue+uint32(c.options.DeadZone) > command.esc {
		returnCommand.esc = MidValue
	}

	if command.esc < MidValue && MidValue-uint32(c.options.DeadZone) < command.esc {
		returnCommand.esc = MidValue
	}

	if command.steer > MidValue && MidValue+uint32(c.options.DeadZone) > command.steer {
		returnCommand.steer = MidValue
	}

	if command.steer < MidValue && MidValue-uint32(c.options.DeadZone) < command.steer {
		returnCommand.steer = MidValue
	}

	if command.pan > MidValue && MidValue+uint32(c.options.DeadZone) > command.pan {
		returnCommand.pan = MidValue
	}

	if command.pan < MidValue && MidValue-uint32(c.options.DeadZone) < command.pan {
		returnCommand.pan = MidValue
	}

	if command.tilt > MidValue && MidValue+uint32(c.options.DeadZone) > command.tilt {
		returnCommand.tilt = MidValue
	}

	if command.tilt < MidValue && MidValue-uint32(c.options.DeadZone) < command.tilt {
		returnCommand.tilt = MidValue
	}

	return returnCommand
}
