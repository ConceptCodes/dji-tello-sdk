# Example: Basic Drone Operations

This example demonstrates how to use the DJI Tello SDK to make the drone take off and then land.

## Prerequisites

- Ensure the drone is powered on and connected to your computer via Wi-Fi.
- Install the DJI Tello SDK Go package: `github.com/conceptcodes/dji-tello-sdk-go`.

## Setup

First, initialize the SDK and create a new drone instance:

```go
package main

import (
  "log"

  "github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
)

func main() {
  // Initialize the SDK
  sdk := tello.NewTelloSDK()

  // Create a new drone instance
  drone, err := sdk.Initialize()
  if err != nil {
    log.Fatalf("Error initializing Tello SDK: %v", err)
    return
  }

  // Proceed with drone operations
  log.Println("Tello SDK initialized successfully!")
}
```

## Take Off

Next, use the `TakeOff` method to make the drone take off:

```go
  // Make the drone take off
  err := drone.TakeOff()
  if err != nil {
    log.Fatalf("Error taking off: %v", err)
    return
  }
  log.Println("Drone is taking off...")
```

## Land
After performing your operations, you can land the drone using the `Land` method:

```go
  // Make the drone land
  err = drone.Land()
  if err != nil {
    log.Fatalf("Error landing: %v", err)
    return
  }
  log.Println("Drone is landing...")
```

## Notes

- Always ensure the drone is in a safe environment before taking off.
- Refer to the [DJI Tello SDK documentation](https://github.com/conceptcodes/dji-tello-sdk-go/README.MD) for additional commands and features.
- Use proper error handling to manage unexpected scenarios during drone operations.
