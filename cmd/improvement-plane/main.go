package main

import (
	"flag"
	"log"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	improvementplane "github.com/piplabs/rsi-agent-platform/internal/improvementplane"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func main() {
	mode := flag.String("mode", "serve", "serve, cron, or worker")
	once := flag.Bool("once", false, "run one cron tick and exit")
	flag.Parse()

	cfg := config.Load("improvement-plane")
	store := storepkg.MustOpenStore(cfg)

	if *mode == "cron" {
		log.Printf("starting %s cron interval=%s", cfg.ServiceName, cfg.ProposalPromoterInterval)
		improvementplane.RunCron(cfg, store, *once)
		return
	}
	if *mode == "worker" {
		log.Printf("starting %s worker poll=%s", cfg.ServiceName, cfg.WorkerPollInterval)
		if err := improvementplane.RunWorker(cfg, store); err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Printf("starting %s on :%d", cfg.ServiceName, cfg.HTTPPort)
	if err := app.ListenAndServe(cfg, improvementplane.NewRouter(cfg, store)); err != nil {
		log.Fatal(err)
	}
}
