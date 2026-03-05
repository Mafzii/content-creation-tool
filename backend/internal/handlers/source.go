package handlers

import (
	"backend/internal/fileutil"
	"backend/internal/models"
	"backend/internal/scraper"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

type SourceStore interface {
	GetAll() ([]models.Source, error)
	Get(id int) (models.Source, error)
	Create(item models.Source) (models.Source, error)
	Update(id int, item models.Source) (models.Source, error)
	Delete(id int) error
}

type SourceHandler struct {
	store SourceStore
}

func NewSourceHandler(store SourceStore) *SourceHandler {
	return &SourceHandler{store: store}
}

// Create handles POST /sources with type-specific logic.
func (h *SourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	var src models.Source

	// Check if this is a multipart file upload
	if len(contentType) >= 19 && contentType[:19] == "multipart/form-data" {
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
			http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
		src.Name = r.FormValue("name")
		src.Type = "file"

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		content, err := fileutil.ReadTextFile(header.Filename, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		src.Raw = header.Filename
		src.Content = content
		src.Status = "ready"
	} else {
		// JSON body
		if err := json.NewDecoder(r.Body).Decode(&src); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
	}

	switch src.Type {
	case "url":
		src.Status = "pending"
		if src.Content == "" {
			src.Content = ""
		}
		created, err := h.store.Create(src)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		// Kick off background fetch
		go h.fetchURLContent(created.Id, src.Raw)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
		return

	case "file":
		// Already handled above in multipart section; status is "ready"

	default:
		// text type
		src.Type = "text"
		if src.Content == "" {
			src.Content = src.Raw
		}
		src.Status = "ready"
	}

	created, err := h.store.Create(src)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// Fetch handles POST /sources/{id}/fetch — re-fetch a URL source.
func (h *SourceHandler) Fetch(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	src, err := h.store.Get(id)
	if err != nil {
		http.Error(w, "source not found", http.StatusNotFound)
		return
	}

	if src.Type != "url" {
		http.Error(w, "only URL sources can be fetched", http.StatusBadRequest)
		return
	}

	// Set to pending
	src.Status = "pending"
	src.Content = ""
	h.store.Update(id, src)

	// Kick off background fetch
	go h.fetchURLContent(id, src.Raw)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "pending"})
}

// Status handles GET /sources/{id}/status — returns just the status.
func (h *SourceHandler) Status(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	src, err := h.store.Get(id)
	if err != nil {
		http.Error(w, "source not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": src.Status})
}

func (h *SourceHandler) fetchURLContent(id int, url string) {
	title, content, err := scraper.FetchURL(url)
	if err != nil {
		slog.Error("failed to fetch URL", "id", id, "url", url, "error", err)
		src, getErr := h.store.Get(id)
		if getErr != nil {
			return
		}
		src.Status = "error"
		h.store.Update(id, src)
		return
	}

	src, err := h.store.Get(id)
	if err != nil {
		return
	}
	src.Content = content
	src.Status = "ready"
	if src.Name == "" && title != "" {
		src.Name = title
	}
	h.store.Update(id, src)
	slog.Info("fetched URL content", "id", id, "url", url, "content_length", len(content))
}
