package handlers_test

import (
	"backend/internal/handlers"
	"backend/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubTopicStore struct {
	topics map[int]models.Topic
}

func (s *stubTopicStore) Get(id int) (models.Topic, error) {
	t, ok := s.topics[id]
	if !ok {
		return models.Topic{}, errors.New("not found")
	}
	return t, nil
}

type stubStyleStore struct {
	styles map[int]models.Style
}

func (s *stubStyleStore) Get(id int) (models.Style, error) {
	st, ok := s.styles[id]
	if !ok {
		return models.Style{}, errors.New("not found")
	}
	return st, nil
}

type stubSourceStore struct {
	sources map[int]models.Source
}

func (s *stubSourceStore) Get(id int) (models.Source, error) {
	src, ok := s.sources[id]
	if !ok {
		return models.Source{}, errors.New("not found")
	}
	return src, nil
}

type stubLLMSettings struct {
	data map[string]string
	err  error
}

func (s *stubLLMSettings) GetAll() (map[string]string, error) {
	return s.data, s.err
}

func (s *stubLLMSettings) Set(key, value string) error { return nil }

func TestGenerateHandler_Generate(t *testing.T) {
	t.Run("rejects missing topic", func(t *testing.T) {
		h := handlers.NewGenerateHandler(
			&stubLLMSettings{data: map[string]string{}},
			&stubTopicStore{topics: map[int]models.Topic{}},
			&stubStyleStore{styles: map[int]models.Style{1: {Id: 1, Name: "formal"}}},
			&stubSourceStore{sources: map[int]models.Source{}},
		)

		body, _ := json.Marshal(map[string]any{"topic_id": 99, "style_id": 1})
		req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewReader(body))
		resp := httptest.NewRecorder()
		h.Generate(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", resp.Code, http.StatusBadRequest)
		}
	})

	t.Run("rejects missing style", func(t *testing.T) {
		h := handlers.NewGenerateHandler(
			&stubLLMSettings{data: map[string]string{}},
			&stubTopicStore{topics: map[int]models.Topic{1: {Id: 1, Name: "Go"}}},
			&stubStyleStore{styles: map[int]models.Style{}},
			&stubSourceStore{sources: map[int]models.Source{}},
		)

		body, _ := json.Marshal(map[string]any{"topic_id": 1, "style_id": 99})
		req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewReader(body))
		resp := httptest.NewRecorder()
		h.Generate(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", resp.Code, http.StatusBadRequest)
		}
	})
}

func TestGenerateHandler_Tweak(t *testing.T) {
	h := handlers.NewGenerateHandler(
		&stubLLMSettings{data: map[string]string{}},
		&stubTopicStore{topics: map[int]models.Topic{}},
		&stubStyleStore{styles: map[int]models.Style{}},
		&stubSourceStore{sources: map[int]models.Source{}},
	)

	body, _ := json.Marshal(map[string]string{"content": "", "instruction": "make shorter"})
	req := httptest.NewRequest(http.MethodPost, "/tweak", bytes.NewReader(body))
	resp := httptest.NewRecorder()
	h.Tweak(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.Code, http.StatusBadRequest)
	}
}
