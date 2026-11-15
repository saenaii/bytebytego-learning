package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bytebytego-learning/design-a-url-shortner/internal/config"
	"github.com/bytebytego-learning/design-a-url-shortner/internal/handler"
	"github.com/bytebytego-learning/design-a-url-shortner/internal/metrics"
	"github.com/bytebytego-learning/design-a-url-shortner/internal/repository"
	"github.com/bytebytego-learning/design-a-url-shortner/internal/service"
	"github.com/bytebytego-learning/design-a-url-shortner/web"
)

func main() {
	configPath := flag.String("config", "config.json", "path to JSON configuration")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := repository.NewPostgres(ctx, cfg.Postgres.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	cache := repository.NewRedis(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB, cfg.TTL())
	defer cache.Close()
	collector := &metrics.Metrics{}
	baseURL := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
	assets, err := fs.Sub(web.FS, "static")
	if err != nil {
		log.Fatal(err)
	}
	svc := service.New(store, cache, collector, strings.TrimRight(baseURL, "/"), cfg.DefaultMode, cfg.QueryBlacklist)
	server := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Server.Port), Handler: handler.New(svc, assets), ReadHeaderTimeout: 5 * time.Second}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	log.Printf("url shortener listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
