package config

import "testing"

func TestLoadPanicsOnInvalidIntegerEnv(t *testing.T) {
	t.Setenv("RSI_HTTP_PORT", "not-a-number")
	assertPanics(t, func() {
		_ = Load("control-plane")
	})
}

func TestLoadPanicsOnInvalidDurationEnv(t *testing.T) {
	t.Setenv("RSI_RUNNER_PROD_TIMEOUT", "not-a-duration")
	assertPanics(t, func() {
		_ = Load("control-plane")
	})
}

func TestLoadPanicsOnInvalidBoolEnv(t *testing.T) {
	t.Setenv("RSI_SLACK_SOCKET_MODE_ENABLED", "maybe")
	assertPanics(t, func() {
		_ = Load("control-plane")
	})
}

func TestLoadPanicsOnInvalidMapEnv(t *testing.T) {
	t.Setenv("RSI_GITHUB_REPO_OWNERS", "story-api")
	assertPanics(t, func() {
		_ = Load("control-plane")
	})
}

func TestLoadPanicsOnInvalidListEnv(t *testing.T) {
	t.Setenv("RSI_ALLOWED_TARGET_REPOS", "depin-backend,,rsi-agent-platform")
	assertPanics(t, func() {
		_ = Load("control-plane")
	})
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}
