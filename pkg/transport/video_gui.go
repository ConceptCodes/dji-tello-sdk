package transport

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/overlay"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

// VideoDisplayType represents the type of video display
type VideoDisplayType string

const (
	DisplayTypeTerminal VideoDisplayType = "terminal"
	DisplayTypeWeb      VideoDisplayType = "web"
)

// VideoDisplay handles real-time video display
type VideoDisplay struct {
	displayType  VideoDisplayType
	frameChan    <-chan VideoFrame
	mlResultChan <-chan ml.MLResult
	isRunning    bool
	mutex        sync.Mutex
	frameCount   int
	startTime    time.Time
	lastFrame    image.Image
	webServer    *http.Server
	webPort      int
	overlay      *overlay.Renderer
	mlConfig     *ml.MLConfig
	lastMLResult map[string]ml.MLResult
	metrics      ml.PipelineMetrics
}

// NewVideoDisplay creates a new video display
func NewVideoDisplay(displayType VideoDisplayType) *VideoDisplay {
	return &VideoDisplay{
		displayType:  displayType,
		webPort:      8080,
		lastMLResult: make(map[string]ml.MLResult),
	}
}

// SetVideoChannel sets the video frame channel
func (vd *VideoDisplay) SetVideoChannel(frameChan <-chan VideoFrame) {
	vd.frameChan = frameChan
}

// SetMLResultChannel sets the ML result channel
func (vd *VideoDisplay) SetMLResultChannel(mlResultChan <-chan ml.MLResult) {
	vd.mlResultChan = mlResultChan
}

// SetMLConfig sets the ML configuration for overlay rendering
func (vd *VideoDisplay) SetMLConfig(mlConfig *ml.MLConfig) {
	vd.mlConfig = mlConfig
	if mlConfig != nil && mlConfig.Overlay.Enabled {
		vd.overlay = overlay.NewRenderer(&mlConfig.Overlay)
	}
}

// SetWebPort sets the web server port for web display
func (vd *VideoDisplay) SetWebPort(port int) {
	vd.webPort = port
}

// Start begins the video display
func (vd *VideoDisplay) Start() error {
	vd.mutex.Lock()
	defer vd.mutex.Unlock()

	if vd.isRunning {
		return fmt.Errorf("video display is already running")
	}

	if vd.frameChan == nil {
		return fmt.Errorf("video frame channel is not set")
	}

	vd.isRunning = true
	vd.startTime = time.Now()
	vd.frameCount = 0

	switch vd.displayType {
	case DisplayTypeTerminal:
		go vd.processTerminalDisplay()
	case DisplayTypeWeb:
		go vd.startWebServer()
		go vd.processWebDisplay()
	default:
		return fmt.Errorf("unsupported display type: %s", vd.displayType)
	}

	// Start ML result processing if channel is available
	if vd.mlResultChan != nil {
		go vd.processMLResults()
	}

	utils.Logger.Infof("Video display started (%s mode)", vd.displayType)
	return nil
}

// Stop stops the video display
func (vd *VideoDisplay) Stop() {
	vd.mutex.Lock()
	defer vd.mutex.Unlock()

	if !vd.isRunning {
		return
	}

	vd.isRunning = false

	// Stop web server if running
	if vd.webServer != nil {
		go func() {
			if err := vd.webServer.Close(); err != nil {
				utils.Logger.Errorf("Error stopping web server: %v", err)
			}
		}()
	}

	utils.Logger.Info("Video display stopped")
}

// IsRunning returns whether the video display is active
func (vd *VideoDisplay) IsRunning() bool {
	vd.mutex.Lock()
	defer vd.mutex.Unlock()
	return vd.isRunning
}

// processTerminalDisplay displays video frames in terminal
func (vd *VideoDisplay) processTerminalDisplay() {
	fmt.Println("🎥 Tello Video Display - Terminal Mode")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println("================================")

	for {
		select {
		case frame, ok := <-vd.frameChan:
			if !ok {
				fmt.Println("\nVideo channel closed")
				return
			}

			if !vd.isRunning {
				return
			}

			vd.frameCount++
			baseImage := vd.createSimpleImage(frame)

			// Apply overlay if available
			if vd.overlay != nil && len(vd.lastMLResult) > 0 {
				// Convert to drawable image
				drawableImg := image.NewRGBA(baseImage.Bounds())
				draw.Draw(drawableImg, baseImage.Bounds(), baseImage, image.Point{}, draw.Src)
				vd.lastFrame = vd.overlay.Render(drawableImg, vd.lastMLResult, vd.metrics)
			} else {
				vd.lastFrame = baseImage
			}

			// Clear screen and display frame info
			vd.displayTerminalFrame(frame)

			// Update every 30 frames
			if vd.frameCount%30 == 0 {
				elapsed := time.Since(vd.startTime)
				fps := float64(vd.frameCount) / elapsed.Seconds()
				fmt.Printf("\n📊 Stats: %d frames | %.1f FPS | %.1fs elapsed\n",
					vd.frameCount, fps, elapsed.Seconds())
			}

		default:
			if !vd.isRunning {
				return
			}
			time.Sleep(33 * time.Millisecond) // ~30 FPS
		}
	}
}

// displayTerminalFrame displays frame information in terminal
func (vd *VideoDisplay) displayTerminalFrame(frame VideoFrame) {
	// Clear screen (ANSI escape code)
	fmt.Print("\033[2J\033[H")

	// Display frame header
	fmt.Printf("🎥 Tello Drone Video Feed\n")
	fmt.Printf("================================\n")
	fmt.Printf("Frame #%d | Size: %d bytes | Keyframe: %v\n",
		frame.SeqNum, frame.Size, frame.IsKeyFrame)
	fmt.Printf("Timestamp: %s\n", frame.Timestamp.Format("15:04:05.000"))
	fmt.Printf("NAL Units: %d\n", len(frame.NALUnits))

	// Display simple ASCII representation
	fmt.Printf("\n📹 Video Frame Visualization:\n")
	vd.displayASCIIArt(frame)

	// Display controls
	fmt.Printf("\n\n🎮 Controls:\n")
	fmt.Printf("• Press Ctrl+C to stop\n")
	fmt.Printf("• Frame rate: ~30 FPS\n")
	fmt.Printf("• Resolution: 960x720\n")
}

// displayASCIIArt creates simple ASCII art representation
func (vd *VideoDisplay) displayASCIIArt(frame VideoFrame) {
	// Create a simple ASCII representation based on frame data
	width := 40
	height := 10

	// Use frame data to create pattern
	pattern := make([][]rune, height)
	for y := 0; y < height; y++ {
		pattern[y] = make([]rune, width)
		for x := 0; x < width; x++ {
			// Create pattern based on frame properties
			if frame.IsKeyFrame {
				if (x+y)%2 == 0 {
					pattern[y][x] = '█'
				} else {
					pattern[y][x] = '▓'
				}
			} else {
				if (x*y)%3 == 0 {
					pattern[y][x] = '▒'
				} else {
					pattern[y][x] = '░'
				}
			}
		}
	}

	// Print ASCII art
	for _, row := range pattern {
		fmt.Printf("  %s\n", string(row))
	}
}

// startWebServer starts the web server for video display
func (vd *VideoDisplay) startWebServer() {
	mux := http.NewServeMux()

	// Serve video stream page
	mux.HandleFunc("/", vd.handleWebPage)
	mux.HandleFunc("/video.jpg", vd.handleVideoFrame)

	// TODO: Integrate modern web server when commander is available
	// This would require adding commander dependency to VideoDisplay

	vd.webServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", vd.webPort),
		Handler: mux,
	}

	fmt.Printf("🌐 Web server started on http://localhost:%d\n", vd.webPort)
	if err := vd.webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		utils.Logger.Errorf("Web server error: %v", err)
	}
}

// handleWebPage serves the HTML page for video display
func (vd *VideoDisplay) handleWebPage(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Tello Video Feed</title>
    <style>
        body { 
            margin: 0; 
            padding: 20px; 
            background: #1a1a1a; 
            color: white; 
            font-family: Arial, sans-serif;
        }
        .container { 
            max-width: 1000px; 
            margin: 0 auto; 
            text-align: center;
        }
        .video-container { 
            background: #000; 
            border: 2px solid #333; 
            margin: 20px 0; 
            position: relative;
        }
        .video-frame { 
            width: 960px; 
            height: 720px; 
            background: #222;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 24px;
        }
        .stats { 
            background: #333; 
            padding: 15px; 
            border-radius: 5px; 
            margin: 10px 0;
        }
        .refresh-btn {
            background: #4CAF50;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
            margin: 10px;
        }
        .refresh-btn:hover {
            background: #45a049;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎥 Tello Drone Video Feed</h1>
        <div class="video-container">
            <div class="video-frame" id="videoFrame">
                Loading video feed...
            </div>
        </div>
        <div class="stats" id="stats">
            Waiting for video data...
        </div>
        <button class="refresh-btn" onclick="refreshFrame()">Refresh Frame</button>
        <button class="refresh-btn" onclick="toggleAutoRefresh()">Toggle Auto-Refresh</button>
    </div>
    
    <script>
        let autoRefresh = true;
        let frameCount = 0;
        
        function refreshFrame() {
            const img = new Image();
            img.onload = function() {
                document.getElementById('videoFrame').innerHTML = '';
                document.getElementById('videoFrame').appendChild(img);
                frameCount++;
                updateStats();
            };
            img.onerror = function() {
                document.getElementById('videoFrame').innerHTML = '⚠️ No video signal';
            };
            img.src = '/video.jpg?t=' + Date.now();
        }
        
        function updateStats() {
            const now = new Date();
            document.getElementById('stats').innerHTML = 
                '📊 Frame: ' + frameCount + ' | Last Update: ' + now.toLocaleTimeString() + 
                ' | Auto-Refresh: ' + (autoRefresh ? 'ON' : 'OFF');
        }
        
        function toggleAutoRefresh() {
            autoRefresh = !autoRefresh;
            if (autoRefresh) {
                startAutoRefresh();
            }
        }
        
        function startAutoRefresh() {
            if (autoRefresh) {
                refreshFrame();
                setTimeout(startAutoRefresh, 100); // ~10 FPS
            }
        }
        
        // Start auto-refresh
        startAutoRefresh();
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleVideoFrame serves the current video frame as JPEG
func (vd *VideoDisplay) handleVideoFrame(w http.ResponseWriter, r *http.Request) {
	vd.mutex.Lock()
	defer vd.mutex.Unlock()

	if vd.lastFrame != nil {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Convert image to JPEG and write response
		if err := png.Encode(w, vd.lastFrame); err != nil {
			utils.Logger.Errorf("Error encoding frame: %v", err)
			http.Error(w, "Frame encoding error", http.StatusInternalServerError)
		}
	} else {
		// Return a placeholder image
		placeholder := vd.createPlaceholderImage()
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, placeholder)
	}
}

// processWebDisplay processes frames for web display
func (vd *VideoDisplay) processWebDisplay() {
	for {
		select {
		case frame, ok := <-vd.frameChan:
			if !ok {
				return
			}

			if !vd.isRunning {
				return
			}

			vd.frameCount++
			baseImage := vd.createSimpleImage(frame)

			// Apply overlay if available
			if vd.overlay != nil && len(vd.lastMLResult) > 0 {
				// Convert to drawable image
				drawableImg := image.NewRGBA(baseImage.Bounds())
				draw.Draw(drawableImg, baseImage.Bounds(), baseImage, image.Point{}, draw.Src)
				vd.lastFrame = vd.overlay.Render(drawableImg, vd.lastMLResult, vd.metrics)
			} else {
				vd.lastFrame = baseImage
			}

		default:
			if !vd.isRunning {
				return
			}
			time.Sleep(33 * time.Millisecond) // ~30 FPS
		}
	}
}

// createSimpleImage creates a simple image from frame data
func (vd *VideoDisplay) createSimpleImage(frame VideoFrame) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 960, 720))

	// Fill with a color based on frame properties
	var c color.RGBA
	if frame.IsKeyFrame {
		c = color.RGBA{R: 100, G: 200, B: 100, A: 255} // Green for keyframes
	} else {
		c = color.RGBA{R: 100, G: 100, B: 200, A: 255} // Blue for other frames
	}

	// Fill image
	for y := 0; y < 720; y++ {
		for x := 0; x < 960; x++ {
			img.Set(x, y, c)
		}
	}

	// Add some pattern based on frame data
	for i := 0; i < len(frame.Data) && i < 1000; i += 10 {
		x := (i * 2) % 960
		y := (i / 2) % 720
		img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	}

	return img
}

// createPlaceholderImage creates a placeholder image
func (vd *VideoDisplay) createPlaceholderImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 960, 720))

	// Fill with dark gray
	for y := 0; y < 720; y++ {
		for x := 0; x < 960; x++ {
			img.Set(x, y, color.RGBA{R: 50, G: 50, B: 50, A: 255})
		}
	}

	return img
}

// GetStats returns current display statistics
func (vd *VideoDisplay) GetStats() map[string]interface{} {
	vd.mutex.Lock()
	defer vd.mutex.Unlock()

	stats := make(map[string]interface{})
	stats["is_running"] = vd.isRunning
	stats["frame_count"] = vd.frameCount
	stats["display_type"] = string(vd.displayType)
	stats["web_port"] = vd.webPort

	if vd.isRunning && !vd.startTime.IsZero() {
		stats["duration"] = time.Since(vd.startTime)
		if vd.frameCount > 0 {
			stats["fps"] = float64(vd.frameCount) / time.Since(vd.startTime).Seconds()
		}
	}

	return stats
}

// processMLResults processes ML results for overlay rendering
func (vd *VideoDisplay) processMLResults() {
	for {
		select {
		case result, ok := <-vd.mlResultChan:
			if !ok {
				utils.Logger.Info("ML result channel closed")
				return
			}

			if !vd.isRunning {
				return
			}

			vd.mutex.Lock()
			// Store result by processor name
			processorName := result.GetProcessorName()
			vd.lastMLResult[processorName] = result
			vd.mutex.Unlock()

			utils.Logger.Debugf("Received ML result from %s processor", processorName)

		default:
			if !vd.isRunning {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// Close stops the display and cleans up resources
func (vd *VideoDisplay) Close() {
	vd.Stop()
}
