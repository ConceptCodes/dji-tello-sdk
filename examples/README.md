# Examples

This directory contains example code and tutorials for using the DJI Tello SDK.

## Available Examples

### [Gamepad Control](./gamepad.md)
Learn how to use a gamepad to control your DJI Tello drone with customizable button mappings and real-time input monitoring.

**Features:**
- Gamepad detection and configuration
- RC value conversion
- Button action mapping
- Real-time state monitoring

### [ML Video Overlay](./ml-video-overlay.md)
Discover how to integrate machine learning processing with real-time video overlay for object detection and tracking.

**Features:**
- YOLO object detection
- Real-time tracking
- Web-based visualization
- Custom overlay rendering

### [Modern Mission Control](./modern-mission-control.md)
Build a modern web-based mission control interface with advanced drone management capabilities.

**Features:**
- Responsive web UI
- Real-time telemetry
- Mission planning
- Advanced controls

### [MP4 Recording](./mp4-recording.md)
Record video from your Tello drone to MP4 format using FFmpeg with customizable recording options.

**Features:**
- H.264 video recording
- FFmpeg integration
- Duration control
- File management

### [Overlay System Test](./overlay-test.md)
Test the ML overlay rendering system with synthetic data to validate functionality.

**Features:**
- Overlay configuration
- Synthetic data testing
- Component validation
- Performance testing

### [Video GUI](./video-gui.md)
Create a simple web-based video GUI for displaying live video from your Tello drone.

**Features:**
- Live video streaming
- Web interface
- FPS overlay
- Responsive design

## Getting Started

### Prerequisites

1. **Go Installation**: Go 1.19 or later
2. **DJI Tello Drone**: Connected to WiFi network
3. **Network Access**: Computer on same network as drone
4. **Optional Dependencies**:
   - FFmpeg for video recording examples
   - Gamepad for control examples
   - ONNX Runtime for ML examples

### Basic Setup

```bash
# Clone the repository
git clone https://github.com/your-repo/dji-tello-sdk.git
cd dji-tello-sdk

# Install dependencies
go mod download

# For ML examples, set library path
export DYLD_LIBRARY_PATH=./lib:$DYLD_LIBRARY_PATH  # macOS
export LD_LIBRARY_PATH=./lib:$LD_LIBRARY_PATH      # Linux

# For recording examples, install FFmpeg
brew install ffmpeg  # macOS
sudo apt install ffmpeg  # Ubuntu
```

### Running Examples

Each example can be run in two ways:

1. **Copy Code**: Copy the code from the markdown file into a `.go` file
2. **Direct Execution**: Use `go run` with inline code (for simple examples)

```bash
# Example: Running the video GUI
go run video_gui_example.go

# Or copy the code and run
cp examples/video-gui.md my_video_gui.go
# Extract the code block and save
go run my_video_gui.go
```

## Example Categories

### 🎮 Control Examples
- **Gamepad Control**: Physical gamepad integration
- **Mission Control**: Advanced web-based control interface

### 📹 Video Examples
- **Video GUI**: Basic video display
- **MP4 Recording**: Video file recording
- **ML Video Overlay**: AI-enhanced video

### 🤖 Machine Learning Examples
- **ML Video Overlay**: Object detection and tracking
- **Overlay System Test**: ML system validation

### 🧪 Testing Examples
- **Overlay System Test**: Component testing and validation

## Configuration

### Common Configuration Files

Examples use configuration files in the `configs/` directory:

- `gamepad-default.json`: Gamepad button mappings
- `ml-pipeline-default.json`: ML processing configuration
- `yolo-default.json`: YOLO model settings

### Environment Variables

```bash
# Set drone IP (default: 192.168.10.1)
export TELLO_IP=192.168.10.1

# Set video port (default: 11111)
export VIDEO_PORT=11111

# Set web interface port (default: 8080)
export WEB_PORT=8080

# For ML examples
export DYLD_LIBRARY_PATH=./lib:$DYLD_LIBRARY_PATH
```

## Troubleshooting

### Common Issues

1. **Drone Connection**
   - Ensure drone is powered on
   - Check WiFi connection to drone network
   - Verify firewall settings

2. **Video Streaming**
   - Check port availability (11111)
   - Ensure sufficient network bandwidth
   - Verify UDP packet forwarding

3. **ML Processing**
   - Install ONNX Runtime library
   - Check model file availability
   - Verify library path configuration

4. **Gamepad Support**
   - Check gamepad connection
   - Verify driver installation
   - Test with different gamepads

### Debug Mode

Enable debug logging for troubleshooting:

```go
// In your Go code
import "github.com/sirupsen/logrus"

utils.Logger.SetLevel(logrus.DebugLevel)
```

### Performance Tips

1. **Network Optimization**
   - Use wired connection when possible
   - Minimize network interference
   - Optimize WiFi channel

2. **System Resources**
   - Close unnecessary applications
   - Monitor CPU and memory usage
   - Use SSD for recording

3. **Video Quality**
   - Adjust resolution based on bandwidth
   - Optimize frame rate for performance
   - Use appropriate compression settings

## Contributing

Have an example to contribute? Please:

1. Create a new markdown file in `examples/`
2. Follow the existing format and structure
3. Include comprehensive documentation
4. Test your code thoroughly
5. Update this index file

## Support

For issues with specific examples:
- Check the example's troubleshooting section
- Review the main SDK documentation
- Open an issue on GitHub
- Join our community discussions

## Additional Resources

- [Main Documentation](../README.md)
- [API Reference](../pkg/)
- [Configuration Guide](../configs/)
- [Troubleshooting Guide](../docs/troubleshooting.md)