package config

import (
	"reflect"
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

	if got := cfg.RunnerTimeoutForRole("prod"); got != 1830*time.Second {
		t.Fatalf("prod runner timeout = %s, want 1830s", got)
	}
	if got := cfg.RunnerTaskTimeoutForRole("prod"); got != 1800*time.Second {
		t.Fatalf("prod runner task timeout = %s, want 1800s", got)
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

func TestLoadUsesModernNotionMirrorDefaults(t *testing.T) {
	t.Setenv("RSI_NOTION_API_VERSION", "")
	t.Setenv("RSI_NOTION_MIRROR_DELTA_ENABLED", "")
	t.Setenv("RSI_NOTION_MIRROR_DELTA_LOOKBACK", "")
	t.Setenv("RSI_NOTION_MIRROR_FULL_SCAN_INTERVAL", "")

	cfg := Load("control-plane")

	if cfg.NotionAPIVersion != "2026-03-11" {
		t.Fatalf("NotionAPIVersion = %q", cfg.NotionAPIVersion)
	}
	if !cfg.NotionMirrorDeltaEnabled {
		t.Fatal("expected notion delta mirror to default enabled")
	}
	if cfg.NotionMirrorDeltaLookback != 10*time.Minute {
		t.Fatalf("delta lookback = %s", cfg.NotionMirrorDeltaLookback)
	}
	if cfg.NotionMirrorFullScanInterval != 24*time.Hour {
		t.Fatalf("full scan interval = %s", cfg.NotionMirrorFullScanInterval)
	}
}

func TestLoadUsesCompanyWikiCompilerRunBudgetDefaults(t *testing.T) {
	t.Setenv("RSI_COMPANY_WIKI_COMPILER_RUN_TIMEOUT", "")
	t.Setenv("RSI_COMPANY_WIKI_COMPILER_SHUTDOWN_GRACE", "")

	cfg := Load("control-plane")

	if cfg.CompanyWikiCompilerRunTimeout != 25*time.Minute {
		t.Fatalf("CompanyWikiCompilerRunTimeout = %s", cfg.CompanyWikiCompilerRunTimeout)
	}
	if cfg.CompanyWikiCompilerShutdownGrace != 30*time.Second {
		t.Fatalf("CompanyWikiCompilerShutdownGrace = %s", cfg.CompanyWikiCompilerShutdownGrace)
	}
}

func TestLoadUsesSourceControlledDBReadApprovers(t *testing.T) {
	t.Setenv("RSI_DB_READ_APPROVER_SLACK_USER_IDS", "")

	cfg := Load("control-plane")

	want := defaultDBReadApproverSlackUserIDs()
	if !reflect.DeepEqual(cfg.DBReadApproverSlackUserIDs, want) {
		t.Fatalf("DBReadApproverSlackUserIDs = %#v, want %#v", cfg.DBReadApproverSlackUserIDs, want)
	}
}

func TestLoadMergesDBReadApproverEnvWithSourceControlledDefaults(t *testing.T) {
	t.Setenv("RSI_DB_READ_APPROVER_SLACK_USER_IDS", "U0772SH7BRA,UEXTRA")

	cfg := Load("control-plane")

	want := append(defaultDBReadApproverSlackUserIDs(), "UEXTRA")
	if !reflect.DeepEqual(cfg.DBReadApproverSlackUserIDs, want) {
		t.Fatalf("DBReadApproverSlackUserIDs = %#v, want %#v", cfg.DBReadApproverSlackUserIDs, want)
	}
}

func TestDefaultDBReadApproversIncludesBlake(t *testing.T) {
	for _, user := range defaultDBReadApproverSlackUsers {
		if user.ID == "U0772SH7BRA" && user.Name == "Blake" {
			return
		}
	}
	t.Fatal("expected Blake to remain in the default DB-read approver allowlist")
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
