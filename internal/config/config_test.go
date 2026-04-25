package config

import (
	"testing"
	"time"
)

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

func TestRunnerTaskTimeoutDefaultsUseExpandedBudgets(t *testing.T) {
	cfg := Config{}

	if got := cfg.RunnerTimeoutForRole("prod"); got != 930*time.Second {
		t.Fatalf("prod runner timeout = %s, want 930s", got)
	}
	if got := cfg.RunnerTaskTimeoutForRole("prod"); got != 900*time.Second {
		t.Fatalf("prod runner task timeout = %s, want 900s", got)
	}
	if got := cfg.RunnerTaskTimeoutForRole("proposal"); got != 420*time.Second {
		t.Fatalf("proposal runner task timeout = %s, want 420s", got)
	}
}

func TestLoadReadsVerboseTraceLoggingEnv(t *testing.T) {
	t.Setenv("RSI_VERBOSE_TRACE_LOGGING", "true")
	t.Setenv("RSI_VERBOSE_TRACE_LOG_LIMIT", "4242")

	cfg := Load("control-plane")

	if !cfg.VerboseTraceLogging {
		t.Fatal("expected verbose trace logging to be enabled")
	}
	if cfg.VerboseTraceLogLimit != 4242 {
		t.Fatalf("verbose trace log limit = %d, want 4242", cfg.VerboseTraceLogLimit)
	}
}

func TestKubernetesReadNamespaceScopeAddsSandboxNamespaceWhenConfigured(t *testing.T) {
	cfg := Config{
		KubernetesReadNamespaces: []string{"story", "rsi-platform", "story"},
		SandboxNamespace:         "rsi-platform",
	}

	got := cfg.KubernetesReadNamespaceScope()
	want := []string{"story", "rsi-platform"}
	if len(got) != len(want) {
		t.Fatalf("scope = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("scope = %#v, want %#v", got, want)
		}
	}
}

func TestKubernetesReadNamespaceScopeUnsetPreservesLegacyUnscopedBehavior(t *testing.T) {
	cfg := Config{SandboxNamespace: "rsi-platform"}

	if got := cfg.KubernetesReadNamespaceScope(); len(got) != 0 {
		t.Fatalf("scope = %#v, want empty", got)
	}
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
