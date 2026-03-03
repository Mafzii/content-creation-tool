package main

import (
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/handlers"
	"backend/internal/middleware"
	"backend/internal/store"
	"log"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg := config.Load()

	database, err := db.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	topicStore := store.NewTopicSQLiteStore(database)
	sourceStore := store.NewSourceSQLiteStore(database)
	styleStore := store.NewStyleSQLiteStore(database)
	draftStore := store.NewDraftSQLiteStore(database)
	settingsStore := store.NewSettingsStore(database)

	topics := handlers.NewCrudHandler(topicStore)
	sourceCrud := handlers.NewCrudHandler(sourceStore)
	sourceHandler := handlers.NewSourceHandler(sourceStore)
	styles := handlers.NewCrudHandler(styleStore)
	drafts := handlers.NewCrudHandler(draftStore)
	settingsHandler := handlers.NewSettingsHandler(settingsStore)
	generateHandler := handlers.NewGenerateHandler(settingsStore, topicStore, styleStore, sourceStore)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /topics", topics.GetAll)
	mux.HandleFunc("GET /topics/{id}", topics.Get)
	mux.HandleFunc("POST /topics", topics.Create)
	mux.HandleFunc("PUT /topics/{id}", topics.Update)
	mux.HandleFunc("DELETE /topics/{id}", topics.Delete)

	mux.HandleFunc("GET /sources", sourceCrud.GetAll)
	mux.HandleFunc("GET /sources/{id}", sourceCrud.Get)
	mux.HandleFunc("POST /sources", sourceHandler.Create)
	mux.HandleFunc("PUT /sources/{id}", sourceCrud.Update)
	mux.HandleFunc("DELETE /sources/{id}", sourceCrud.Delete)
	mux.HandleFunc("POST /sources/{id}/fetch", sourceHandler.Fetch)
	mux.HandleFunc("GET /sources/{id}/status", sourceHandler.Status)

	mux.HandleFunc("GET /styles", styles.GetAll)
	mux.HandleFunc("GET /styles/{id}", styles.Get)
	mux.HandleFunc("POST /styles", styles.Create)
	mux.HandleFunc("PUT /styles/{id}", styles.Update)
	mux.HandleFunc("DELETE /styles/{id}", styles.Delete)

	mux.HandleFunc("GET /drafts", drafts.GetAll)
	mux.HandleFunc("GET /drafts/{id}", drafts.Get)
	mux.HandleFunc("POST /drafts", drafts.Create)
	mux.HandleFunc("PUT /drafts/{id}", drafts.Update)
	mux.HandleFunc("DELETE /drafts/{id}", drafts.Delete)

	mux.HandleFunc("GET /settings", settingsHandler.GetAll)
	mux.HandleFunc("POST /settings", settingsHandler.Set)

	mux.HandleFunc("POST /generate", generateHandler.Generate)
	mux.HandleFunc("POST /tweak", generateHandler.Tweak)

	frontendDir := "../frontend"
	if dir := os.Getenv("FRONTEND_DIR"); dir != "" {
		frontendDir = dir
	}
	mux.Handle("/", http.FileServer(http.Dir(frontendDir)))

	slog.Info("server starting", "port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, middleware.Logging(mux)))
}
