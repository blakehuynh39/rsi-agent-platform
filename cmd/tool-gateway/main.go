package main

import (
	"log"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	platformdb "github.com/piplabs/rsi-agent-platform/internal/db"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/toolgateway"
)

func main() {
	cfg, err := config.Load("tool-gateway").ValidatedFor("tool-gateway", "serve")
	if err != nil {
		log.Fatal(err)
	}
	store := storepkg.MustOpenStore(cfg)
	if provider, ok := store.(interface {
		SchemaStatus() platformdb.SchemaStatus
	}); ok {
		status := provider.SchemaStatus()
		cfg.SchemaVersionCurrent = status.CurrentVersion
		cfg.SchemaVersionExpected = status.ExpectedVersion
		cfg.SchemaCompatibility = status.State
	}
	log.Printf("starting %s kind=%s mode=%s on :%d dependencies=%v", cfg.ServiceName, cfg.ServiceKind, cfg.RuntimeMode, cfg.HTTPPort, cfg.DependencyTargets())
	if err := app.ListenAndServe(cfg, toolgateway.NewRouter(cfg, store)); err != nil {
		log.Fatal(err)
	}
}
