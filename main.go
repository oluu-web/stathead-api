package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"stathead/internal/handler"
	"stathead/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	playerStore := store.NewPlayerStore(pool)
	playerHandler := handler.NewPlayerHandler(playerStore)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{os.Getenv("ALLOWED_ORIGIN")},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/players", func(r chi.Router) {
		r.Get("/search",               playerHandler.Search)
		r.Get("/compare",              playerHandler.Compare)
		r.Get("/leaders",              playerHandler.Leaders)
		r.Get("/{playerID}/seasons",   playerHandler.Seasons)
		r.Get("/{playerID}/gamelogs",  playerHandler.GameLogs)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("server starting", "port", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
