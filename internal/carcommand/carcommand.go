package carcommand

import (
	"context"
	"log"
	"sync"
	"time"
)

const midpoint = byte(127)

type CarCommand struct {
	name           string
	refreshRate    int
	CommandChannel chan Command
	LatestCommand  LatestCommand
}

type LatestCommand struct {
	lock    sync.RWMutex
	used    bool
	command Command
}

type Command struct {
	esc   byte
	servo byte
	pan   byte
	tilt  byte
}

func NewCarCommand(name string, refreshRate int) *CarCommand {
	return &CarCommand{
		name:           name,
		refreshRate:    refreshRate,
		CommandChannel: make(chan Command, 5),
		LatestCommand: LatestCommand{
			command: Command{
				esc:   midpoint,
				servo: midpoint,
				pan:   midpoint,
				tilt:  midpoint,
			},
		},
	}
}

func (c *CarCommand) Start(ctx context.Context) {
	go func() {
		commandRate := 1000 / c.refreshRate
		commandDuration := time.Duration(int64(time.Millisecond) * int64(commandRate))
		commandTicker := time.NewTicker(commandDuration)
		for {
			select {
			case <-ctx.Done():
				log.Printf("car command stopped: %s\n", ctx.Err())
				return
			case command, ok := <-c.CommandChannel:
				if !ok {
					log.Println("car command channel stopped")
					return
				}
				c.LatestCommand.lock.Lock()
				c.LatestCommand.used = false
				c.LatestCommand.command = command
				c.LatestCommand.lock.Unlock()
			case <-commandTicker.C:
				c.LatestCommand.lock.RLock()
				if c.LatestCommand.used {
					c.LatestCommand.lock.RUnlock()
					log.Printf("command was already used, skipping")
					continue
				}
				c.LatestCommand.lock.RUnlock()

				c.LatestCommand.lock.Lock()
				c.LatestCommand.used = true
				//c.sendCommand(command)
				c.LatestCommand.lock.Unlock()
			}
		}
	}()
}
