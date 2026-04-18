package main

import (
	"flag"
	"log"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	platformdb "github.com/piplabs/rsi-agent-platform/internal/db"
	improvementplane "github.com/piplabs/rsi-agent-platform/internal/improvementplane"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func main() {
	mode := flag.String("mode", "serve", "serve, cron, reconcile, worker, or migrate")
	once := flag.Bool("once", false, "run one cron/reconcile tick and exit")
	flag.Parse()

	cfg, err := config.Load("improvement-plane").ValidatedFor("improvement-plane", *mode)
	if err != nil {
		log.Fatal(err)
	}
	if *mode == "migrate" {
		db, err := platformdb.OpenPostgres(cfg.PostgresURL)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		status, err := platformdb.ApplyMigrations(db)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("completed %s kind=%s mode=%s schema=%d/%d state=%s", cfg.ServiceName, cfg.ServiceKind, cfg.RuntimeMode, status.CurrentVersion, status.ExpectedVersion, status.State)
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
	if *mode == "reconcile" {
		log.Printf("starting %s kind=%s mode=%s poll=%s dependencies=%v", cfg.ServiceName, cfg.ServiceKind, cfg.RuntimeMode, cfg.WorkerPollInterval, cfg.DependencyTargets())
		if *once {
			if err := improvementplane.RunReconcilePass(cfg, store); err != nil {
				log.Fatal(err)
			}
			return
		}
		if err := improvementplane.RunReconciler(cfg, store); err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Printf("starting %s kind=%s mode=%s on :%d dependencies=%v", cfg.ServiceName, cfg.ServiceKind, cfg.RuntimeMode, cfg.HTTPPort, cfg.DependencyTargets())
	if err := app.ListenAndServe(cfg, improvementplane.NewRouter(cfg, store)); err != nil {
		log.Fatal(err)
	}
}
