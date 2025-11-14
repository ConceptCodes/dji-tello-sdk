package web

import (
	"net/http"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

// EnhancedVideoDisplay extends existing VideoDisplay with modern web capabilities
type EnhancedVideoDisplay struct {
	videoDisplay *transport.VideoDisplay
	webServer    *WebServer
}

// NewEnhancedVideoDisplay creates a new enhanced video display with modern web UI
func NewEnhancedVideoDisplay(commander tello.TelloCommander, mlResultChan <-chan ml.MLResult) *EnhancedVideoDisplay {
	// Create base video display
	baseDisplay := transport.NewVideoDisplay(transport.DisplayTypeWeb)

	// Create web server with proper ML result channel
	webServer := NewWebServer(commander, mlResultChan)

	return &EnhancedVideoDisplay{
		videoDisplay: baseDisplay,
		webServer:    webServer,
	}
}

// StartEnhanced starts enhanced video display with modern web UI
func (evd *EnhancedVideoDisplay) StartEnhanced() error {
	// Set up enhanced web routes
	mux := http.NewServeMux()
	evd.webServer.SetupRoutes(mux)

	// Override video frame handler to work with our mux
	// mux.HandleFunc("/video.jpg", evd.videoDisplay.GetVideoFrameHandler())

	// Serve modern index page instead of basic one
	mux.HandleFunc("/", evd.webServer.handleIndex)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start web server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Logger.Errorf("Web server error: %v", err)
		}
	}()

	return nil
}
