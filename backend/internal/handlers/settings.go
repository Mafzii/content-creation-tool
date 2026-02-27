package handlers

import (
	"encoding/json"
	"net/http"
)

type settingsStoreIface interface {
	GetAll() (map[string]string, error)
	Set(key, value string) error
}

type SettingsHandler struct {
	store settingsStoreIface
}

func NewSettingsHandler(s settingsStoreIface) *SettingsHandler {
	return &SettingsHandler{store: s}
}

func (h *SettingsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	settings, err := h.store.GetAll()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (h *SettingsHandler) Set(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.store.Set(body.Key, body.Value); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
