package handlers

import (
	"backend/internal/llm"
	"backend/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type topicStoreIface interface {
	Get(id int) (models.Topic, error)
}

type styleStoreIface interface {
	Get(id int) (models.Style, error)
}

type sourceStoreIface interface {
	Get(id int) (models.Source, error)
}

type GenerateHandler struct {
	settings settingsStoreIface
	topics   topicStoreIface
	styles   styleStoreIface
	sources  sourceStoreIface
}

func NewGenerateHandler(
	settings settingsStoreIface,
	topics topicStoreIface,
	styles styleStoreIface,
	sources sourceStoreIface,
) *GenerateHandler {
	return &GenerateHandler{
		settings: settings,
		topics:   topics,
		styles:   styles,
		sources:  sources,
	}
}

func (h *GenerateHandler) buildClient() (llm.Client, error) {
	return BuildLLMClient(h.settings)
}

// BuildLLMClient creates an LLM client from stored settings.
func BuildLLMClient(settings settingsStoreIface) (llm.Client, error) {
	cfg, err := settings.GetAll()
	if err != nil {
		return nil, err
	}
	provider := cfg["llm_provider"]
	model := cfg["llm_model"]
	apiKey := cfg["llm_api_key"]
	if provider == "" {
		provider = "gemini"
	}
	return llm.NewClient(provider, model, apiKey)
}

func (h *GenerateHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TopicId   int    `json:"topic_id"`
		StyleId   int    `json:"style_id"`
		SourceIds []int  `json:"source_ids"`
		Notes     string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	topic, err := h.topics.Get(req.TopicId)
	if err != nil {
		http.Error(w, "topic not found", http.StatusBadRequest)
		return
	}
	style, err := h.styles.Get(req.StyleId)
	if err != nil {
		http.Error(w, "style not found", http.StatusBadRequest)
		return
	}

	var sourceParts []string
	for _, sid := range req.SourceIds {
		src, err := h.sources.Get(sid)
		if err != nil {
			continue
		}
		text := src.Content
		if text == "" {
			text = src.Raw
		}
		if text != "" {
			sourceParts = append(sourceParts, fmt.Sprintf("- %s: %s", src.Name, text))
		}
	}

	prompt := buildPrompt(topic, style, sourceParts, req.Notes)

	client, err := h.buildClient()
	if err != nil {
		http.Error(w, "LLM config error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	variants, err := client.Generate(r.Context(), prompt)
	if err != nil {
		http.Error(w, "generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"variants": variants})
}

func (h *GenerateHandler) Tweak(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content     string `json:"content"`
		Instruction string `json:"instruction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	client, err := h.buildClient()
	if err != nil {
		http.Error(w, "LLM config error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := client.Tweak(r.Context(), req.Content, req.Instruction)
	if err != nil {
		http.Error(w, "tweak failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"content": result})
}

func buildPrompt(topic models.Topic, style models.Style, sourceParts []string, notes string) string {
	var sb strings.Builder
	sb.WriteString("You are a content writer. Write in the following style:\n")
	if style.Tone != "" {
		sb.WriteString("Tone: " + style.Tone + "\n")
	}
	if style.Prompt != "" {
		sb.WriteString(style.Prompt + "\n")
	}
	if style.Example != "" {
		sb.WriteString("\nExample of this style:\n" + style.Example + "\n")
	}
	sb.WriteString("\nTopic: " + topic.Name + "\n")
	if topic.Description != "" {
		sb.WriteString(topic.Description + "\n")
	}
	if topic.Keywords != "" {
		sb.WriteString("Keywords to include: " + topic.Keywords + "\n")
	}
	if len(sourceParts) > 0 {
		sb.WriteString("\nReference material:\n")
		for _, p := range sourceParts {
			sb.WriteString(p + "\n")
		}
	}
	if notes != "" {
		sb.WriteString("\nAdditional instructions: " + notes + "\n")
	}
	sb.WriteString("\nWrite a complete draft with a title.")
	return sb.String()
}
