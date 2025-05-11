package main

import (
	"context"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
)

func main() {
	// Create a new Tello instance
	ctx := context.Background()

	commandConn, err := transport.NewCommandConn(ctx, 8889, 5*time.Second)
	if err != nil {
		panic(err)
	}
	stateConn, err := transport.NewConn(ctx, 5*time.Second)
	if err != nil {
		panic(err)
	}

	stateStream := transport.NewStateStream(ctx, stateConn)

	drone := tello.InitializeSDK(ctx, commandConn, stateStream, nil)
	drone.TakeOff()
}
