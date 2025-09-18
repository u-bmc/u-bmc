# KVM Service

The KVM service provides Keyboard, Video, Mouse functionality for BMC environments, enabling remote console access to managed systems.

## Overview

The KVM service integrates:
- **Video Capture**: Captures video frames from V4L2 devices (`/dev/video0`)
- **USB Gadget Emulation**: Provides HID keyboard and mouse emulation via Linux USB gadget subsystem
- **VNC Server**: Serves video and accepts input via VNC protocol on port 5900
- **HTTP Streaming**: Optional MJPEG streaming on port 8080
- **Mass Storage**: Optional virtual media mounting

## Features

- **Multiple Protocols**: VNC (TCP + WebSocket) and HTTP MJPEG streaming
- **USB HID**: Full keyboard and mouse emulation with proper HID descriptors
- **Virtual Media**: Mass storage emulation for ISO mounting and virtual disks
- **Configurable Quality**: Adjustable video quality and frame rates
- **Multi-Client**: Support for multiple concurrent VNC and HTTP clients
- **Resource Management**: Automatic cleanup and graceful shutdown
- **Error Recovery**: Built-in retry mechanisms and fault tolerance

## Configuration

### Basic Usage

```go
service := kvmsrv.New(
    kvmsrv.WithConfig(kvmsrv.DefaultConfig()),
    kvmsrv.WithName("kvm-server"),
)
```

### Custom Configuration

```go
config := &kvmsrv.Config{
    VideoDevice:       "/dev/video0",
    VideoWidth:        1920,
    VideoHeight:       1080,
    VideoFPS:          30,
    VNCPort:           5900,
    HTTPPort:          8080,
    EnableVNC:         true,
    EnableHTTP:        true,
    EnableUSB:         true,
    EnableMassStorage: true,
    VNCPassword:       "secure123",
    VNCMaxClients:     3,
    HTTPMaxClients:    5,
    JPEGQuality:       85,
}

service := kvmsrv.New(
    kvmsrv.WithConfig(config),
    kvmsrv.WithName("kvm-hd"),
)
```

### Example Configurations

The service provides several pre-configured examples:

```go
// High-resolution setup
config := kvmsrv.ExampleHighResConfig()

// Security-focused setup
config := kvmsrv.ExampleSecureConfig()

// Low-bandwidth optimized
config := kvmsrv.ExampleLowBandwidthConfig()

// By use case
config := kvmsrv.ExampleConfigByUseCase("production")
```

## API Reference

### Service Interface

```go
type Service interface {
    Name() string
    Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `VideoDevice` | string | `/dev/video0` | V4L2 video capture device |
| `VideoWidth` | uint32 | `640` | Capture width in pixels |
| `VideoHeight` | uint32 | `480` | Capture height in pixels |
| `VideoFPS` | uint32 | `30` | Target frames per second |
| `VNCPort` | int | `5900` | VNC server port |
| `VNCWebSocketPort` | int | `5901` | VNC WebSocket port |
| `HTTPPort` | int | `8080` | HTTP streaming port |
| `EnableVNC` | bool | `true` | Enable VNC server |
| `EnableHTTP` | bool | `true` | Enable HTTP streaming |
| `EnableUSB` | bool | `true` | Enable USB gadget |
| `EnableMassStorage` | bool | `true` | Enable virtual media |
| `VNCPassword` | string | `""` | VNC authentication password |
| `VNCMaxClients` | int | `5` | Maximum VNC clients |
| `HTTPMaxClients` | int | `10` | Maximum HTTP clients |
| `JPEGQuality` | int | `75` | JPEG compression quality (1-100) |
| `USBGadgetName` | string | `"kvm-gadget"` | USB gadget identifier |
| `USBVendorID` | string | `"0x1d6b"` | USB vendor ID |
| `USBProductID` | string | `"0x0104"` | USB product ID |
| `ClientTimeout` | Duration | `30m` | Client idle timeout |
| `FrameTimeout` | Duration | `5s` | Video capture timeout |
| `BufferCount` | uint32 | `4` | Video capture buffers |

### USB HID Functions

```go
// Send keyboard input
err := service.SendKeyboardInput(ctx, modifier, keys)

// Send mouse movement (absolute coordinates)
err := service.SendMouseInput(ctx, x, y, buttons)

// Send mouse wheel
err := service.SendWheelInput(ctx, wheel)

// Mount virtual media
err := service.SetMassStorageFile(ctx, "/path/to/image.iso", true)
```

### Status Monitoring

```go
// Get service status
status := service.GetStatus()

// Check if running
if service.IsRunning() {
    // Service is active
}
```

## Network Protocols

### VNC Protocol

- **Port**: 5900 (TCP)
- **WebSocket**: 5901 (HTTP/WebSocket)
- **Authentication**: Optional password
- **Encodings**: Raw, RRE, Hextile
- **Pixel Formats**: 32-bit RGBA

### HTTP Streaming

- **Port**: 8080 (HTTP)
- **Format**: MJPEG (multipart/x-mixed-replace)
- **Path**: `/stream` for video, `/` for viewer
- **Quality**: Configurable JPEG compression

## USB Gadget Integration

### HID Devices

- **Keyboard**: `/dev/hidg0` - Standard USB HID keyboard
- **Mouse**: `/dev/hidg1` - Absolute positioning mouse with wheel
- **Report Descriptors**: Full USB HID compliance

### Mass Storage

- **Device**: USB mass storage LUN 0
- **Modes**: CD-ROM or removable disk
- **Features**: Hot-swappable virtual media

### Gadget Management

The service automatically:
- Creates USB composite gadget
- Configures HID functions
- Binds to available UDC
- Handles gadget lifecycle
- Recovers from failures

## Error Handling

The service provides comprehensive error handling:

```go
// Service-specific errors
var (
    ErrServiceNotConfigured       = errors.New("KVM service not configured")
    ErrVideoDeviceUnavailable     = errors.New("video capture device unavailable")
    ErrUSBGadgetInitFailed        = errors.New("USB gadget initialization failed")
    ErrVNCServerFailed            = errors.New("VNC server failed")
    ErrHTTPServerFailed           = errors.New("HTTP streaming server failed")
    ErrInvalidConfiguration       = errors.New("invalid service configuration")
    ErrResourceUnavailable        = errors.New("required resource unavailable")
    ErrOperationFailed            = errors.New("KVM operation failed")
    ErrTimeout                    = errors.New("operation timed out")
    ErrInvalidFrame               = errors.New("invalid video frame")
)
```

## Performance Considerations

### Video Capture

- **Buffer Management**: Configurable capture buffers (2-8 recommended)
- **Frame Rate**: Balance between quality and CPU usage
- **Resolution**: Higher resolutions require more bandwidth and processing
- **Format**: YUYV preferred for capture, MJPEG for compressed sources

### Network Streaming

- **VNC**: More efficient for static content, supports compression
- **HTTP**: Better for simple viewers, higher bandwidth usage
- **Client Limits**: Prevent resource exhaustion with reasonable limits
- **Quality Settings**: JPEG quality 60-85 provides good balance

### USB Performance

- **HID Latency**: Sub-10ms input response times
- **Mass Storage**: Virtual media performance depends on backing storage
- **Gadget Binding**: Recovery mechanisms handle UDC disconnections

## Deployment

### System Requirements

- Linux kernel with V4L2 support
- USB gadget support (configfs)
- Video capture device (`/dev/video0`)
- Appropriate permissions for:
  - `/dev/video*` (video group)
  - `/sys/kernel/config` (root or usbgadget group)
  - `/dev/hidg*` (root or input group)

### Service Integration

```go
// In your BMC service manager
kvmService := kvmsrv.New(
    kvmsrv.WithConfig(config),
    kvmsrv.WithName("kvm"),
)

// Register and start
serviceManager.Register(kvmService)
```

### Monitoring

The service provides status information:

```go
status := kvmService.GetStatus()
fmt.Printf("Video: %dx%d @ %dfps, %d frames captured\n",
    status.VideoCapture.Width,
    status.VideoCapture.Height,
    status.VideoCapture.FPS,
    status.VideoCapture.Frames)

fmt.Printf("VNC: %d/%d clients connected\n",
    status.VNCServer.Clients,
    status.VNCServer.MaxClients)
```

## Troubleshooting

### Common Issues

1. **Video Device Not Found**
   - Check `/dev/video0` exists and is accessible
   - Verify V4L2 driver is loaded
   - Check permissions (`ls -l /dev/video*`)

2. **USB Gadget Failed**
   - Ensure configfs is mounted (`mount | grep configfs`)
   - Check UDC availability (`ls /sys/class/udc/`)
   - Verify permissions on `/sys/kernel/config`

3. **VNC Connection Issues**
   - Check port 5900 is not blocked by firewall
   - Verify service is listening (`netstat -tlnp | grep 5900`)
   - Test with VNC client (e.g., `vncviewer localhost:5900`)

4. **Poor Video Quality**
   - Increase JPEG quality setting
   - Check capture device supports requested resolution
   - Monitor CPU usage during streaming

### Debug Mode

Enable debug logging for detailed troubleshooting:

```go
// Service logs detailed status information
// Check system logs for KVM service messages
```

## License

This software is licensed under the BSD-3-Clause license.