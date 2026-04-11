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
	mode := flag.String("mode", "serve", "serve, slack-surface, or worker")
	flag.Parse()

	cfg := config.Load("control-plane")
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
	log.Printf("starting %s mode=%s on :%d", cfg.ServiceName, *mode, cfg.HTTPPort)
	if err := app.ListenAndServe(cfg, control.NewRouter(cfg, store)); err != nil {
		log.Fatal(err)
	}
}
