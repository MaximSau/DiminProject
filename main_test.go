package main

import (
	"os"
	"testing"
)

func TestEnvOr(t *testing.T) {
	const key = "TEST_PULSEFIT_ENV_VAR"
	os.Unsetenv(key)

	if val := envOr(key, "default_val"); val != "default_val" {
		t.Errorf("expected default_val, got %s", val)
	}

	os.Setenv(key, "custom_val")
	defer os.Unsetenv(key)

	if val := envOr(key, "default_val"); val != "custom_val" {
		t.Errorf("expected custom_val, got %s", val)
	}
}

func TestSetupLogger(t *testing.T) {
	// Test text handler with debug level
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "text")
	setupLogger()

	// Test json handler in prod environment
	os.Setenv("ENV", "production")
	os.Setenv("PULSEFIT_LOG_FORMAT", "json")
	setupLogger()

	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LOG_FORMAT")
	os.Unsetenv("ENV")
	os.Unsetenv("PULSEFIT_LOG_FORMAT")
	setupLogger()
}
