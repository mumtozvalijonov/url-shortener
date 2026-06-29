package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mumtozvalijonov/url-shortener/config"
	"github.com/mumtozvalijonov/url-shortener/internal/adapters/generator"
	"github.com/mumtozvalijonov/url-shortener/internal/adapters/httpapi"
	"github.com/mumtozvalijonov/url-shortener/internal/adapters/persistence/postgres"
	"github.com/mumtozvalijonov/url-shortener/internal/services"
)

func main() {
	config, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	router := http.NewServeMux()
	db, err := sql.Open("pgx", config.PostgresDSN)
	if err != nil {
		log.Fatalf("sql open: %v", err)
	}
	shortURLRepository := postgres.NewShortURLRepository(db)
	generator := generator.NewRandomGenerator(5)
	shortenerService := services.NewShortenerService(shortURLRepository, generator)
	handler := httpapi.NewHandler(shortenerService)
	handler.RegisterRoutes(router)

	server := http.Server{
		Addr:              config.HTTP.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen and serve: %v", err)
		}
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-shutdownCtx.Done()
	stop()

	log.Println("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}

	log.Println("server stopped")
}
