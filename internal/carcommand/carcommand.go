package carcommand

import (
	"context"
	"fmt"
	"log"
	"time"
)

type CarCommand struct {
	CommandChannel chan CommandGroup

	config          CarCommandConfig
	servoController *ServoController
}

type CarCommandConfig struct {
	RefreshRate           int
	ServoControllerConfig ServoControllerConfig
	ServoConfigs          []ServoConfig
}

type CommandGroup struct {
	Commands map[string]Command
}

type Command struct {
	Value int
	Gear  string
}

func NewCarCommand(cfg CarCommandConfig) *CarCommand {
	carCommand := CarCommand{
		CommandChannel:  make(chan CommandGroup, 5),
		servoController: NewServoController(cfg.ServoControllerConfig),
		config:          cfg,
	}
	return &carCommand
}

// Make connection over I2C and start creating servos
func (c *CarCommand) Init() error {
	err := c.servoController.Init()
	if err != nil {
		return fmt.Errorf("failed initializing servo controller - %w", err)
	}

	log.Printf("Adding Servos...\n\n")
	for _, servoCfg := range c.config.ServoConfigs {
		c.servoController.AddServo(servoCfg)
	}
	return nil
}

func (c *CarCommand) Start(ctx context.Context) error {
	c.Init()

	commandRate := 1000 / c.config.RefreshRate
	commandDuration := time.Duration(int64(time.Millisecond) * int64(commandRate))
	commandTicker := time.NewTicker(commandDuration)

	var latestCommand CommandGroup
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("car command stopped: %s\n", ctx.Err())

		case command, ok := <-c.CommandChannel: //recieved command from client
			if !ok {
				return fmt.Errorf("car command channel stopped")
			}
			latestCommand = command //Use this command on next cycle

		case <-commandTicker.C: //time to send command
			if latestCommand.Commands != nil {
				err := c.DoCommand(latestCommand)
				if err != nil {
					return err
				}
				latestCommand.Commands = nil
			}
		}
	}
}

func (c *CarCommand) DoCommand(commands CommandGroup) error {
	for i, command := range commands.Commands {
		err := c.servoController.SendCommand(i, int(command.Value))
		if err != nil {
			return fmt.Errorf("error sending command (name: %s | command %d) - %w", err)
		}

		err = c.servoController.SetGear(i, command.Gear)
		if err != nil {
			return fmt.Errorf("error sending command (name: %s | command %d) - %w", err)
		}
	}
	return nil
}
