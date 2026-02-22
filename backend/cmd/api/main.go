package main

import (
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/handlers"
	"backend/internal/store"
	"context"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	topicStore := store.NewTopicPostgresStore(pool)
	sourceStore := store.NewSourcePostgresStore(pool)
	styleStore := store.NewStylePostgresStore(pool)
	draftStore := store.NewDraftPostgresStore(pool)

	topics := handlers.NewCrudHandler(topicStore)
	sources := handlers.NewCrudHandler(sourceStore)
	styles := handlers.NewCrudHandler(styleStore)
	drafts := handlers.NewCrudHandler(draftStore)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /topics", topics.GetAll)
	mux.HandleFunc("GET /topics/{id}", topics.Get)
	mux.HandleFunc("POST /topics", topics.Create)
	mux.HandleFunc("DELETE /topics/{id}", topics.Delete)

	mux.HandleFunc("GET /sources", sources.GetAll)
	mux.HandleFunc("GET /sources/{id}", sources.Get)
	mux.HandleFunc("POST /sources", sources.Create)
	mux.HandleFunc("DELETE /sources/{id}", sources.Delete)

	mux.HandleFunc("GET /styles", styles.GetAll)
	mux.HandleFunc("GET /styles/{id}", styles.Get)
	mux.HandleFunc("POST /styles", styles.Create)
	mux.HandleFunc("DELETE /styles/{id}", styles.Delete)

	mux.HandleFunc("GET /drafts", drafts.GetAll)
	mux.HandleFunc("GET /drafts/{id}", drafts.Get)
	mux.HandleFunc("POST /drafts", drafts.Create)
	mux.HandleFunc("DELETE /drafts/{id}", drafts.Delete)

	log.Printf("server starting on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
