package main

import (
	"log"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/platform"
)

func main() {
	cfg := config.Load("control-plane")
	store := platform.NewMemoryStore()
	log.Printf("starting %s on :%d", cfg.ServiceName, cfg.HTTPPort)
	if err := platform.ListenAndServe(cfg, platform.NewControlPlaneRouter(cfg, store)); err != nil {
		log.Fatal(err)
	}
}

