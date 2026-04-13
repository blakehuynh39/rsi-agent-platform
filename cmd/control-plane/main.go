package main

import (
	"flag"
	"log"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/control"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func main() {
	mode := flag.String("mode", "serve", "serve, slack-surface, worker, or action-worker")
	flag.Parse()

	cfg, err := config.Load("control-plane").ValidatedFor("control-plane", *mode)
	if err != nil {
		log.Fatal(err)
	}
	store := storepkg.MustOpenStore(cfg)
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
