package handlers_test

import (
	"backend/internal/handlers"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubSettingsStore struct {
	data map[string]string
	err  error
}

func (s *stubSettingsStore) GetAll() (map[string]string, error) {
	return s.data, s.err
}

func (s *stubSettingsStore) Set(key, value string) error {
	if s.err != nil {
		return s.err
	}
	s.data[key] = value
	return nil
}

func TestSettingsHandler_GetAll(t *testing.T) {
	store := &stubSettingsStore{data: map[string]string{"llm_provider": "gemini"}}
	h := handlers.NewSettingsHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	resp := httptest.NewRecorder()
	h.GetAll(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.Code, http.StatusOK)
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	if result["llm_provider"] != "gemini" {
		t.Errorf("llm_provider = %q, want %q", result["llm_provider"], "gemini")
	}
}

func TestSettingsHandler_Set(t *testing.T) {
	t.Run("sets a setting", func(t *testing.T) {
		store := &stubSettingsStore{data: map[string]string{}}
		h := handlers.NewSettingsHandler(store)

		body, _ := json.Marshal(map[string]string{"key": "llm_provider", "value": "claude"})
		req := httptest.NewRequest(http.MethodPost, "/settings", bytes.NewReader(body))
		resp := httptest.NewRecorder()
		h.Set(resp, req)

		if resp.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", resp.Code, http.StatusNoContent)
		}
		if store.data["llm_provider"] != "claude" {
			t.Errorf("stored value = %q, want %q", store.data["llm_provider"], "claude")
		}
	})

	t.Run("rejects empty key", func(t *testing.T) {
		store := &stubSettingsStore{data: map[string]string{}}
		h := handlers.NewSettingsHandler(store)

		body, _ := json.Marshal(map[string]string{"key": "", "value": "test"})
		req := httptest.NewRequest(http.MethodPost, "/settings", bytes.NewReader(body))
		resp := httptest.NewRecorder()
		h.Set(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", resp.Code, http.StatusBadRequest)
		}
	})
}
