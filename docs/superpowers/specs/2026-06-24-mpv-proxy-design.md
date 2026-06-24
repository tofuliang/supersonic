# MPV Proxy Configuration Design

## Overview

**Goal:** Add HTTP/HTTPS proxy support to the MPV player in Supersonic.

**Scope:** Only modify MPV player; DLNA player proxy support is out of scope for this change.

**Key Features:**
- Support HTTP/HTTPS proxy with authentication (username:password in URL)
- Configuration file takes precedence over environment variables
- Backward compatible with existing behavior

## Configuration

### New Config Field

Add `HTTPProxy` field to `LocalPlaybackConfig` in `backend/config.go`:

```go
type LocalPlaybackConfig struct {
    AudioDeviceName       string
    AudioExclusive        bool
    InMemoryCacheSizeMB   int
    Volume                int
    EqualizerEnabled      bool
    EqualizerType         string
    EqualizerPreamp       float64
    GraphicEqualizerBands []float64
    ActiveEQPresetName    string
    AutoEQProfilePath     string
    AutoEQProfileName     string
    PauseFade             bool
    HTTPProxy             string  // NEW: HTTP/HTTPS proxy URL with optional auth
}
```

**Default Value:** Empty string (no proxy)

**Config File Example:**
```toml
[LocalPlayback]
HTTPProxy = "http://user:pass@proxy:8080"
```

## Implementation

### 1. MPV Player Changes

**File:** `backend/player/mpv/player.go`

Modify `Init` function signature:

```go
func (p *Player) Init(maxCacheMB int, httpProxy string) error {
    if !p.initialized {
        m := mpv.Create()

        // ... existing options ...

        // Set proxy if provided
        if httpProxy != "" {
            m.SetOptionString("http-proxy", httpProxy)
        }

        // ... rest of initialization ...
    }
    // ...
}
```

### 2. Application Layer Integration

**File:** `backend/app.go`

Modify `initMPV` function:

```go
func (a *App) initMPV() error {
    p := mpv.NewWithClientName(a.appName)
    c := a.Config.LocalPlayback
    c.InMemoryCacheSizeMB = clamp(c.InMemoryCacheSizeMB, 10, 500)
    
    // Get proxy config: config file first, environment variable as fallback
    httpProxy := c.HTTPProxy
    if httpProxy == "" {
        httpProxy = os.Getenv("https_proxy")
        if httpProxy == "" {
            httpProxy = os.Getenv("HTTPS_PROXY")
        }
    }
    
    if err := p.Init(c.InMemoryCacheSizeMB, httpProxy); err != nil {
        return fmt.Errorf("failed to initialize mpv player: %s", err.Error())
    }
    a.LocalPlayer = p
    return nil
}
```

### 3. Environment Variable Support

**Supported Environment Variables:**
- `https_proxy` (lowercase)
- `HTTPS_PROXY` (uppercase)

**Priority:**
1. Configuration file `HTTPProxy` field
2. `https_proxy` environment variable
3. `HTTPS_PROXY` environment variable
4. No proxy (default)

## Error Handling

### Proxy URL Validation

- MPV validates proxy URL format during initialization
- Invalid URL causes application startup failure with error message

### Proxy Connection Validation

- MPV validates proxy connection when actually streaming
- Connection failure results in playback error visible to user

### No Proxy Scenario

- If `HTTPProxy` is empty and no environment variable is set, MPV doesn't set proxy option
- Behavior identical to current implementation (backward compatible)

### Logging

Recommend logging proxy setting for debugging:

```go
if httpProxy != "" {
    log.Printf("Setting MPV proxy: %s", httpProxy)
}
```

## Testing Plan

### Unit Tests

- Test config field serialization/deserialization
- Test environment variable reading logic (config precedence)

### Integration Tests

- Test MPV initialization with proxy setting
- Test backward compatibility without proxy

### Manual Testing

- Test playback with local proxy server
- Test proxy with authentication
- Test config file vs environment variable precedence

**Test Commands:**

```bash
# Test without proxy
go run .

# Test with config file proxy
# Set HTTPProxy in config file, then run application

# Test with environment variable proxy
HTTPS_PROXY=http://proxy:8080 go run .
```

## Backward Compatibility

- Existing configurations without `HTTPProxy` field continue to work
- No proxy is used by default (same as current behavior)
- No breaking changes to existing API or behavior

## Future Considerations

- SOCKS5 proxy support (if needed in future)
- Runtime proxy configuration changes (if needed in future)
- DLNA player proxy support (separate feature request)