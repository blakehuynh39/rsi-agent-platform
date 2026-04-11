package main

import (
	"log"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/toolgateway"
)

func main() {
	cfg := config.Load("tool-gateway")
	store := storepkg.MustOpenStore(cfg)
	log.Printf("starting %s on :%d", cfg.ServiceName, cfg.HTTPPort)
	if err := app.ListenAndServe(cfg, toolgateway.NewRouter(cfg, store)); err != nil {
		log.Fatal(err)
	}
}
