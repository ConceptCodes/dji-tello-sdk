package tello

import (
	"context"
	"strconv"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

type TelloCommanderConfig struct {
	commChannel  transport.CommandConn
	commandQueue *CommandQueue
}

type TelloCommander interface {
	TakeOff()
	Land()
	GoForward(distance int)
	GoBackward(distance int)
	GoLeft(distance int)
	GoRight(distance int)
	GoUp(distance int)
	GoDown(distance int)
}

func NewTelloCommander(
	ctx context.Context,
	commChannel transport.CommandConn,
	commandQueue *CommandQueue) *TelloCommanderConfig {

	return &TelloCommanderConfig{
		commChannel:  commChannel,
		commandQueue: commandQueue,
	}
}

func (c *TelloCommanderConfig) StartQueue() {
	c.commandQueue.Start(c)
}

func (c *TelloCommanderConfig) Stop() {
	c.commandQueue.Stop()
}

func (c *TelloCommanderConfig) StartCommandListener(commandCh <-chan string) {
	go func() {
		for cmd := range commandCh {
			if err := c.commChannel.Send([]byte(cmd)); err != nil {
				utils.Logger.Errorf("Failed to send command '%s': %v", cmd, err)
			}
		}
	}()
}

func (c *TelloCommanderConfig) TakeOff() {
	c.commandQueue.Enqueue("takeoff")
}

func (c *TelloCommanderConfig) Land() {
	c.commandQueue.Enqueue("land")
}

func (c *TelloCommanderConfig) GoForward(distance int) {
	cmd := "forward " + strconv.Itoa(distance)
	c.commandQueue.Enqueue(cmd)
}

func (c *TelloCommanderConfig) GoBackward(distance int) {
	cmd := "backward " + strconv.Itoa(distance)
	c.commandQueue.Enqueue(cmd)
}

func (c *TelloCommanderConfig) GoLeft(distance int) {
	cmd := "left " + strconv.Itoa(distance)
	c.commandQueue.Enqueue(cmd)
}

func (c *TelloCommanderConfig) GoRight(distance int) {
	cmd := "right " + strconv.Itoa(distance)
	c.commandQueue.Enqueue(cmd)
}

func (c *TelloCommanderConfig) GoUp(distance int) {
	cmd := "up " + strconv.Itoa(distance)
	c.commandQueue.Enqueue(cmd)
}

func (c *TelloCommanderConfig) GoDown(distance int) {
	cmd := "down " + strconv.Itoa(distance)
	c.commandQueue.Enqueue(cmd)
}
