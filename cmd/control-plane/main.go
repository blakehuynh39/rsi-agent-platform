package main

import (
	"context"
	"flag"
	"log"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/control"
	platformdb "github.com/piplabs/rsi-agent-platform/internal/db"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func main() {
	mode := flag.String("mode", "serve", "serve, slack-surface, slack-mirror, worker, or action-worker")
	flag.Parse()

	cfg, err := config.Load("control-plane").ValidatedFor("control-plane", *mode)
	if err != nil {
		log.Fatal(err)
	}
	if *mode == "slack-mirror" {
		mirrorStore, err := storepkg.OpenSourceMirrorWriteStore(cfg)
		if err != nil {
			log.Fatal(err)
		}
		if provider, ok := mirrorStore.(interface {
			SchemaStatus() platformdb.SchemaStatus
		}); ok {
			status := provider.SchemaStatus()
			cfg.SchemaVersionCurrent = status.CurrentVersion
			cfg.SchemaVersionExpected = status.ExpectedVersion
			cfg.SchemaCompatibility = status.State
		}
		if err := control.RunSlackMirror(context.Background(), cfg, mirrorStore); err != nil {
			log.Fatal(err)
		}
		return
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
	if *mode == "slack-surface" {
		if err := control.RunSlackSurface(cfg, store); err != nil {
			log.Fatal(err)
		}
		return
	}
	if *mode == "worker" {
		if err := control.RunWorker(cfg, store); err != nil {
			log.Fatal(err)
		}
		return
	}
	if *mode == "action-worker" {
		if err := control.RunActionWorker(cfg, store); err != nil {
			log.Fatal(err)
		}
		return
	}
	log.Printf("starting %s kind=%s mode=%s on :%d dependencies=%v", cfg.ServiceName, cfg.ServiceKind, cfg.RuntimeMode, cfg.HTTPPort, cfg.DependencyTargets())
	if err := app.ListenAndServe(cfg, control.NewRouter(cfg, store)); err != nil {
		log.Fatal(err)
	}
}
