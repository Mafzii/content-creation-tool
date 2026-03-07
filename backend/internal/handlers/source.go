package handlers

import (
	"backend/internal/fileutil"
	"backend/internal/models"
	"backend/internal/scraper"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

const extractionPrompt = `Extract and summarize the key information from this content. Focus on facts, data, quotes, and arguments. Remove boilerplate and irrelevant content. Return a clean, well-structured summary.`

const topicExtractionPrompt = `Extract and summarize the key information from this content that is relevant to "%s". %sFocus on facts, data, quotes, and arguments related to this subject. Remove boilerplate and irrelevant content. Return a clean, well-structured summary.`

type SourceStore interface {
	GetAll() ([]models.Source, error)
	Get(id int) (models.Source, error)
	Create(item models.Source) (models.Source, error)
	Update(id int, item models.Source) (models.Source, error)
	Delete(id int) error
}

type SourceHandler struct {
	store    SourceStore
	settings settingsStoreIface
	topics   topicStoreIface
}

func NewSourceHandler(store SourceStore, settings settingsStoreIface, topics topicStoreIface) *SourceHandler {
	return &SourceHandler{store: store, settings: settings, topics: topics}
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

	if src.ExtractMode == "" {
		src.ExtractMode = "standard"
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
		go h.fetchURLContent(created.Id, src.Raw, src.ExtractMode, src.TopicId)

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
	go h.fetchURLContent(id, src.Raw, src.ExtractMode, src.TopicId)

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

func (h *SourceHandler) fetchURLContent(id int, url, extractMode string, topicId int) {
	title, rawContent, err := scraper.FetchURL(url)
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

	content := rawContent
	status := "ready"

	if extractMode == "ai" {
		aiContent, aiErr := h.aiExtract(rawContent, topicId)
		if aiErr != nil {
			slog.Error("AI extraction failed, falling back to raw content", "id", id, "error", aiErr)
			status = "partial"
		} else {
			content = aiContent
		}
	}

	src.Content = content
	src.Status = status
	if src.Name == "" && title != "" {
		src.Name = title
	}
	h.store.Update(id, src)
	slog.Info("fetched URL content", "id", id, "url", url, "extract_mode", extractMode, "status", status, "content_length", len(content))
}

func (h *SourceHandler) aiExtract(rawContent string, topicId int) (string, error) {
	client, err := BuildLLMClient(h.settings)
	if err != nil {
		return "", fmt.Errorf("build LLM client: %w", err)
	}

	prompt := extractionPrompt
	if topicId > 0 {
		topic, err := h.topics.Get(topicId)
		if err == nil {
			keywords := ""
			if topic.Keywords != "" {
				keywords = fmt.Sprintf("Focus on these keywords: %s. ", topic.Keywords)
			}
			prompt = fmt.Sprintf(topicExtractionPrompt, topic.Name, keywords)
		}
	}

	result, err := client.Tweak(context.Background(), rawContent, prompt)
	if err != nil {
		return "", fmt.Errorf("LLM tweak: %w", err)
	}
	return result, nil
}
