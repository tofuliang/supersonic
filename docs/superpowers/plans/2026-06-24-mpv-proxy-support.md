# MPV Proxy Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add HTTP/HTTPS proxy support to the MPV player with config file priority over environment variables.

**Architecture:** Modify `LocalPlaybackConfig` to include `HTTPProxy` field, update MPV player initialization to accept proxy parameter, and integrate proxy configuration in the application layer.

**Tech Stack:** Go, MPV (libmpv), TOML configuration

---

## File Structure

### Files to Modify
- `backend/config.go` - Add `HTTPProxy` field to `LocalPlaybackConfig`
- `backend/player/mpv/player.go` - Update `Init` function signature to accept proxy
- `backend/app.go` - Read proxy config and pass to MPV initialization

### Files to Create
- `docs/superpowers/plans/2026-06-24-mpv-proxy-support.md` - This implementation plan

---

## Task 1: Add HTTPProxy Field to Configuration

**Files:**
- Modify: `backend/config.go:133-146`

- [ ] **Step 1: Add HTTPProxy field to LocalPlaybackConfig struct**

```go
type LocalPlaybackConfig struct {
	AudioDeviceName       string
	AudioExclusive        bool
	InMemoryCacheSizeMB   int
	Volume                int
	EqualizerEnabled      bool
	EqualizerType         string    // "ISO10Band" or "ISO15Band"
	EqualizerPreamp       float64
	GraphicEqualizerBands []float64
	ActiveEQPresetName    string // Name of currently selected EQ preset
	AutoEQProfilePath     string // Path to applied AutoEQ profile (e.g., "oratory1990/over-ear/Sennheiser HD 650")
	AutoEQProfileName     string // Display name of applied profile (e.g., "Sennheiser HD 650")
	PauseFade             bool
	HTTPProxy             string  // NEW: HTTP/HTTPS proxy URL with optional auth (http://user:pass@proxy:8080)
}
```

- [ ] **Step 2: Verify the change compiles**

Run: `go build ./backend/...`
Expected: BUILD SUCCESSFUL

- [ ] **Step 3: Commit the configuration change**

```bash
git add backend/config.go
git commit -m "feat: add HTTPProxy field to LocalPlaybackConfig"
```

---

## Task 2: Update MPV Player Init Function

**Files:**
- Modify: `backend/player/mpv/player.go:103-148`

- [ ] **Step 1: Update Init function signature to accept httpProxy parameter**

```go
func (p *Player) Init(maxCacheMB int, httpProxy string) error {
	if !p.initialized {
		m := mpv.Create()

		m.SetOptionString("idle", "yes")
		m.SetOptionString("video", "no")
		m.SetOptionString("audio-display", "no")
		m.SetOptionString("gapless-audio", "weak")
		m.SetOptionString("prefetch-playlist", "yes")
		m.SetOptionString("force-seekable", "yes")
		m.SetOptionString("terminal", "no")

		// limit in-memory cache size
		maxBackMB := maxCacheMB / 3
		maxForwardMB := maxBackMB + maxBackMB
		m.SetOptionString("demuxer-max-bytes", fmt.Sprintf("%dMiB", maxForwardMB))
		m.SetOptionString("demuxer-max-back-bytes", fmt.Sprintf("%dMiB", maxBackMB))

		if p.vol < 0 {
			p.vol = 100
		}
		m.SetOption("volume", mpv.FORMAT_INT64, p.vol)

		p.SetAudioExclusive(p.audioExclusive)
		if p.haveRGainOpts {
			p.SetReplayGainOptions(p.replayGainOpts)
		}

		if p.clientName != "" {
			m.SetOptionString("audio-client-name", p.clientName)
		}

		// Set proxy if provided
		if httpProxy != "" {
			m.SetOptionString("http-proxy", httpProxy)
		}

		m.ObserveProperty(0, "metadata", mpv.FORMAT_NODE)

		if err := m.Initialize(); err != nil {
			return fmt.Errorf("error initializing mpv: %s", err.Error())
		}

		p.mpv = m
	}
	ctx, cancel := context.WithCancel(context.Background())
	go p.eventHandler(ctx)
	p.bgCancel = cancel
	p.initialized = true
	return nil
}
```

- [ ] **Step 2: Verify the change compiles**

Run: `go build ./backend/player/mpv/...`
Expected: BUILD SUCCESSFUL

- [ ] **Step 3: Commit the MPV player change**

```bash
git add backend/player/mpv/player.go
git commit -m "feat: update MPV Init to accept httpProxy parameter"
```

---

## Task 3: Update Application Layer to Pass Proxy

**Files:**
- Modify: `backend/app.go:369-378`

- [ ] **Step 1: Update initMPV function to read and pass proxy config**

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

- [ ] **Step 2: Add import for os package if not already present**

Check if `os` is already imported in `backend/app.go`. If not, add it:

```go
import (
	// ... existing imports ...
	"os"
)
```

- [ ] **Step 3: Verify the change compiles**

Run: `go build ./backend/...`
Expected: BUILD SUCCESSFUL

- [ ] **Step 4: Commit the application layer change**

```bash
git add backend/app.go
git commit -m "feat: integrate proxy configuration in app initialization"
```

---

## Task 4: Add Proxy Logging

**Files:**
- Modify: `backend/app.go:369-378` (same function as Task 3)

- [ ] **Step 1: Add logging when proxy is configured**

Update the `initMPV` function to include logging:

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
	
	// Log proxy configuration for debugging
	if httpProxy != "" {
		log.Printf("Setting MPV proxy: %s", httpProxy)
	}
	
	if err := p.Init(c.InMemoryCacheSizeMB, httpProxy); err != nil {
		return fmt.Errorf("failed to initialize mpv player: %s", err.Error())
	}
	a.LocalPlayer = p
	return nil
}
```

- [ ] **Step 2: Verify the change compiles**

Run: `go build ./backend/...`
Expected: BUILD SUCCESSFUL

- [ ] **Step 3: Commit the logging change**

```bash
git add backend/app.go
git commit -m "feat: add proxy configuration logging"
```

---

## Task 5: Test Configuration Serialization

**Files:**
- Create: `backend/config_test.go`

- [ ] **Step 1: Write test for HTTPProxy field serialization**

```go
package backend

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHTTPProxyConfigSerialization(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	// Write config with HTTPProxy
	configContent := `
[LocalPlayback]
HTTPProxy = "http://user:pass@proxy:8080"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Read and parse config
	cfg, err := ReadConfigFile(configPath, "test")
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Verify HTTPProxy field
	if cfg.LocalPlayback.HTTPProxy != "http://user:pass@proxy:8080" {
		t.Errorf("Expected HTTPProxy 'http://user:pass@proxy:8080', got '%s'", cfg.LocalPlayback.HTTPProxy)
	}
}

func TestHTTPProxyConfigDefault(t *testing.T) {
	// Create a temporary config file without HTTPProxy
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	// Write config without HTTPProxy
	configContent := `
[LocalPlayback]
Volume = 50
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Read and parse config
	cfg, err := ReadConfigFile(configPath, "test")
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Verify HTTPProxy field is empty
	if cfg.LocalPlayback.HTTPProxy != "" {
		t.Errorf("Expected empty HTTPProxy, got '%s'", cfg.LocalPlayback.HTTPProxy)
	}
}
```

- [ ] **Step 2: Run the tests**

Run: `go test ./backend/ -run TestHTTPProxy -v`
Expected: PASS

- [ ] **Step 3: Commit the test**

```bash
git add backend/config_test.go
git commit -m "test: add HTTPProxy configuration serialization tests"
```

---

## Task 6: Test Environment Variable Reading

**Files:**
- Create: `backend/app_test.go`

- [ ] **Step 1: Write test for environment variable fallback**

```go
package backend

import (
	"os"
	"testing"
)

func TestProxyEnvironmentVariableFallback(t *testing.T) {
	// Save original environment
	origHTTPSProxy := os.Getenv("https_proxy")
	origHTTPSPROXY := os.Getenv("HTTPS_PROXY")
	defer func() {
		os.Setenv("https_proxy", origHTTPSProxy)
		os.Setenv("HTTPS_PROXY", origHTTPSPROXY)
	}()

	// Test case 1: No proxy set
	os.Unsetenv("https_proxy")
	os.Unsetenv("HTTPS_PROXY")
	
	// Create a minimal app config
	cfg := &Config{
		LocalPlayback: LocalPlaybackConfig{},
	}
	
	// Simulate the proxy reading logic
	httpProxy := cfg.LocalPlayback.HTTPProxy
	if httpProxy == "" {
		httpProxy = os.Getenv("https_proxy")
		if httpProxy == "" {
			httpProxy = os.Getenv("HTTPS_PROXY")
		}
	}
	
	if httpProxy != "" {
		t.Errorf("Expected empty proxy, got '%s'", httpProxy)
	}

	// Test case 2: Set https_proxy
	os.Setenv("https_proxy", "http://proxy:8080")
	httpProxy = cfg.LocalPlayback.HTTPProxy
	if httpProxy == "" {
		httpProxy = os.Getenv("https_proxy")
		if httpProxy == "" {
			httpProxy = os.Getenv("HTTPS_PROXY")
		}
	}
	
	if httpProxy != "http://proxy:8080" {
		t.Errorf("Expected 'http://proxy:8080', got '%s'", httpProxy)
	}

	// Test case 3: Set HTTPS_PROXY (uppercase)
	os.Unsetenv("https_proxy")
	os.Setenv("HTTPS_PROXY", "http://proxy2:8080")
	httpProxy = cfg.LocalPlayback.HTTPProxy
	if httpProxy == "" {
		httpProxy = os.Getenv("https_proxy")
		if httpProxy == "" {
			httpProxy = os.Getenv("HTTPS_PROXY")
		}
	}
	
	if httpProxy != "http://proxy2:8080" {
		t.Errorf("Expected 'http://proxy2:8080', got '%s'", httpProxy)
	}

	// Test case 4: Config file takes precedence
	cfg.LocalPlayback.HTTPProxy = "http://config-proxy:8080"
	httpProxy = cfg.LocalPlayback.HTTPProxy
	if httpProxy == "" {
		httpProxy = os.Getenv("https_proxy")
		if httpProxy == "" {
			httpProxy = os.Getenv("HTTPS_PROXY")
		}
	}
	
	if httpProxy != "http://config-proxy:8080" {
		t.Errorf("Expected 'http://config-proxy:8080', got '%s'", httpProxy)
	}
}
```

- [ ] **Step 2: Run the tests**

Run: `go test ./backend/ -run TestProxyEnvironmentVariable -v`
Expected: PASS

- [ ] **Step 3: Commit the test**

```bash
git add backend/app_test.go
git commit -m "test: add environment variable fallback tests"
```

---

## Task 7: Manual Verification

**Files:**
- None (manual testing)

- [ ] **Step 1: Test without proxy (backward compatibility)**

Run: `go run .`
Expected: Application starts normally, no proxy configured

- [ ] **Step 2: Test with config file proxy**

1. Edit config file (usually `~/.config/supersonic/config.toml`)
2. Add under `[LocalPlayback]`:
   ```toml
   HTTPProxy = "http://your-proxy:8080"
   ```
3. Run: `go run .`
4. Check logs for: `Setting MPV proxy: http://your-proxy:8080`
5. Try playing a song

- [ ] **Step 3: Test with environment variable**

Run: `HTTPS_PROXY=http://your-proxy:8080 go run .`
Check logs for: `Setting MPV proxy: http://your-proxy:8080`

- [ ] **Step 4: Test config file precedence**

1. Set `HTTPS_PROXY=http://env-proxy:8080` environment variable
2. Set `HTTPProxy = "http://config-proxy:8080"` in config file
3. Run: `go run .`
4. Check logs for: `Setting MPV proxy: http://config-proxy:8080` (config wins)

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "feat: complete MPV proxy support implementation"
```

---

## Summary

This plan implements HTTP/HTTPS proxy support for the MPV player in Supersonic with the following features:

1. **Configuration**: New `HTTPProxy` field in `LocalPlaybackConfig`
2. **Priority**: Config file > `https_proxy` env var > `HTTPS_PROXY` env var > No proxy
3. **Authentication**: Supports `http://user:pass@proxy:8080` format
4. **Logging**: Logs proxy configuration for debugging
5. **Backward Compatible**: Existing configurations continue to work without changes

**Total Tasks:** 7
**Estimated Time:** 30-45 minutes
**Lines Changed:** ~50 lines of code, ~100 lines of tests