package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"


	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"example.com/go_practics/internal/config"
	"example.com/go_practics/internal/storage/postgres"
	httptransport "example.com/go_practics/internal/http"
	rediscache "example.com/go_practics/internal/storage/redis"
)

func main() {
	_ = godotenv.Load()

	cfg := config.FromEnv()

	pgxCfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	pgxCfg.MaxConns = 20
	pgxCfg.MinConns = 5
	pgxCfg.MaxConnLifetime = time.Hour
	pgxCfg.ConnConfig.StatementCacheCapacity = 256

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	repo := postgres.NewRepo(pool)

	// Redis cache
	cache, err := rediscache.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.CacheTTL)
	if err != nil {
		log.Fatal(err)
	}
	defer cache.Close()

	srv := httptransport.NewServer(repo, cache)

	log.Printf("listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, srv.Router()); err != nil {
		log.Fatal(err)
	}
}