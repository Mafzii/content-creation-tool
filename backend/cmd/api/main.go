package main

import (
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/handlers"
	"backend/internal/store"
	"log"
	"net/http"
	"os"
)

func main() {
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

	frontendDir := "../frontend"
	if dir := os.Getenv("FRONTEND_DIR"); dir != "" {
		frontendDir = dir
	}
	mux.Handle("/", http.FileServer(http.Dir(frontendDir)))

	log.Printf("server starting on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
