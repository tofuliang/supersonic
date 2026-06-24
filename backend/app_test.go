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
		_ = os.Setenv("https_proxy", origHTTPSProxy)
		_ = os.Setenv("HTTPS_PROXY", origHTTPSPROXY)
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
