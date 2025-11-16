package main

import (
	"Avito/pkg/config"
	"Avito/pkg/gateway"
	"Avito/pkg/usecase"
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
)

var envFile = flag.String("env-file", ".env", "path to env file")

func main() {
	flag.Parse()
	cfg, err := config.Load(*envFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting Service...")
	log.Printf("Server Port: %d", cfg.Server.Port)
	log.Printf("Database URL: %s", cfg.DBUrl())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pgCfg, err := cfg.PGXConfig()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	cases := usecase.Setup(cfg, pool)

	s := gateway.NewServer(ctx, cfg, cases)
	if err := s.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server error: %v", err)
	}
}
