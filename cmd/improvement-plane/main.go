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

	cfg, err := config.Load("improvement-plane").ValidatedFor("improvement-plane", *mode)
	if err != nil {
		log.Fatal(err)
	}
	store := storepkg.MustOpenStore(cfg)

	if *mode == "cron" {
		log.Printf("starting %s kind=%s mode=%s interval=%s dependencies=%v", cfg.ServiceName, cfg.ServiceKind, cfg.RuntimeMode, cfg.ProposalPromoterInterval, cfg.DependencyTargets())
		improvementplane.RunCron(cfg, store, *once)
		return
	}
	if *mode == "worker" {
		log.Printf("starting %s kind=%s mode=%s poll=%s dependencies=%v", cfg.ServiceName, cfg.ServiceKind, cfg.RuntimeMode, cfg.WorkerPollInterval, cfg.DependencyTargets())
		if err := improvementplane.RunWorker(cfg, store); err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Printf("starting %s kind=%s mode=%s on :%d dependencies=%v", cfg.ServiceName, cfg.ServiceKind, cfg.RuntimeMode, cfg.HTTPPort, cfg.DependencyTargets())
	if err := app.ListenAndServe(cfg, improvementplane.NewRouter(cfg, store)); err != nil {
		log.Fatal(err)
	}
}
