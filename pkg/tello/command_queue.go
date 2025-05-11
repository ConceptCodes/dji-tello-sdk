package tello

import (
	"context"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

type CommandQueue struct {
	ctx       context.Context
	cancel    context.CancelFunc
	commandCh chan string
	wg        sync.WaitGroup
}

func NewCommandQueue(ctx context.Context) *CommandQueue {
	ctx, cancel := context.WithCancel(ctx)
	return &CommandQueue{
		ctx:       ctx,
		cancel:    cancel,
		commandCh: make(chan string, 32),
	}
}

func (cq *CommandQueue) Start(commander *TelloCommanderConfig) {
	cq.wg.Add(1)
	go cq.processCommands(commander)
}

func (cq *CommandQueue) Stop() {
	cq.cancel()
	cq.wg.Wait()
	close(cq.commandCh)
}

func (cq *CommandQueue) Enqueue(command string) {
	select {
	case cq.commandCh <- command:
		utils.Logger.Infof("Command queued: %s", command)
	case <-cq.ctx.Done():
		utils.Logger.Warn("Command queue stopped, unable to enqueue command")
	}
}

func (cq *CommandQueue) processCommands(commander *TelloCommanderConfig) {
	defer cq.wg.Done()

	ticker := time.NewTicker(1 * time.Second) // Enforce 1 Hz rate limit
	defer ticker.Stop()

	for {
		select {
		case <-cq.ctx.Done():
			return
		case command := <-cq.commandCh:
			<-ticker.C
			if err := commander.commChannel.Send([]byte(command)); err != nil {
				utils.Logger.Errorf("Failed to send command '%s': %v", command, err)
			} else {
				utils.Logger.Infof("Command sent: %s", command)
			}
		}
	}
}
