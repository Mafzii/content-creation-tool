package handlers_test

import (
	"backend/internal/handlers"
	"backend/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type stubSourceCRUDStore struct {
	mu      sync.Mutex
	items   map[int]models.Source
	nextID  int
	err     error
	updated map[int]models.Source
}

func newStubSourceCRUDStore() *stubSourceCRUDStore {
	return &stubSourceCRUDStore{
		items:   map[int]models.Source{},
		nextID:  1,
		updated: map[int]models.Source{},
	}
}

func (s *stubSourceCRUDStore) GetAll() ([]models.Source, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return nil, s.err
}

func (s *stubSourceCRUDStore) Get(id int) (models.Source, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return models.Source{}, s.err
	}
	src, ok := s.items[id]
	if !ok {
		return models.Source{}, errors.New("not found")
	}
	return src, nil
}

func (s *stubSourceCRUDStore) Create(item models.Source) (models.Source, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return models.Source{}, s.err
	}
	item.Id = s.nextID
	s.nextID++
	s.items[item.Id] = item
	return item, nil
}

// Update applies the change to s.items and records the first update per ID in
// s.updated. Only the first update is recorded so that the synchronous handler
// update (e.g. setting "pending" status) is captured rather than any subsequent
// write from a background goroutine.
func (s *stubSourceCRUDStore) Update(id int, item models.Source) (models.Source, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return models.Source{}, s.err
	}
	s.items[id] = item
	if _, alreadyRecorded := s.updated[id]; !alreadyRecorded {
		s.updated[id] = item
	}
	return item, nil
}

func (s *stubSourceCRUDStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// getUpdated safely returns the first-recorded update for the given ID.
func (s *stubSourceCRUDStore) getUpdated(id int) models.Source {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.updated[id]
}

func makeSourceHandler(store *stubSourceCRUDStore) *handlers.SourceHandler {
	return handlers.NewSourceHandler(store)
}

func TestSourceHandler_Create(t *testing.T) {
	t.Run("text source copies raw to content when content empty", func(t *testing.T) {
		store := newStubSourceCRUDStore()
		h := makeSourceHandler(store)

		body, _ := json.Marshal(map[string]string{
			"name": "my notes",
			"type": "text",
			"raw":  "some raw content",
		})
		req := httptest.NewRequest(http.MethodPost, "/sources", bytes.NewReader(body))
		resp := httptest.NewRecorder()
		h.Create(resp, req)

		if resp.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", resp.Code, http.StatusCreated)
		}

		var created models.Source
		json.NewDecoder(resp.Body).Decode(&created)
		if created.Status != "ready" {
			t.Errorf("status = %q, want %q", created.Status, "ready")
		}
		if created.Content != "some raw content" {
			t.Errorf("content = %q, want raw content copied", created.Content)
		}
	})

	t.Run("text source preserves explicit content", func(t *testing.T) {
		store := newStubSourceCRUDStore()
		h := makeSourceHandler(store)

		body, _ := json.Marshal(map[string]string{
			"name":    "notes",
			"type":    "text",
			"raw":     "raw",
			"content": "cleaned",
		})
		req := httptest.NewRequest(http.MethodPost, "/sources", bytes.NewReader(body))
		resp := httptest.NewRecorder()
		h.Create(resp, req)

		var created models.Source
		json.NewDecoder(resp.Body).Decode(&created)
		if created.Content != "cleaned" {
			t.Errorf("content = %q, want %q", created.Content, "cleaned")
		}
	})

	t.Run("url source gets pending status", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "<html><body><p>Test content</p></body></html>")
		}))
		defer ts.Close()

		store := newStubSourceCRUDStore()
		h := makeSourceHandler(store)

		body, _ := json.Marshal(map[string]string{
			"name": "wiki",
			"type": "url",
			"raw":  ts.URL,
		})
		req := httptest.NewRequest(http.MethodPost, "/sources", bytes.NewReader(body))
		resp := httptest.NewRecorder()
		h.Create(resp, req)

		if resp.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", resp.Code, http.StatusCreated)
		}

		var created models.Source
		json.NewDecoder(resp.Body).Decode(&created)
		if created.Status != "pending" {
			t.Errorf("status = %q, want %q", created.Status, "pending")
		}
	})

	t.Run("unknown type defaults to text", func(t *testing.T) {
		store := newStubSourceCRUDStore()
		h := makeSourceHandler(store)

		body, _ := json.Marshal(map[string]string{"name": "src", "raw": "data"})
		req := httptest.NewRequest(http.MethodPost, "/sources", bytes.NewReader(body))
		resp := httptest.NewRecorder()
		h.Create(resp, req)

		if resp.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", resp.Code, http.StatusCreated)
		}
		if store.items[1].Type != "text" {
			t.Errorf("type = %q, want %q", store.items[1].Type, "text")
		}
	})
}

func TestSourceHandler_Status(t *testing.T) {
	store := newStubSourceCRUDStore()
	store.items[1] = models.Source{Id: 1, Status: "pending"}
	h := makeSourceHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/sources/1/status", nil)
	req.SetPathValue("id", "1")
	resp := httptest.NewRecorder()
	h.Status(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	if result["status"] != "pending" {
		t.Errorf("status = %q, want %q", result["status"], "pending")
	}
}

func TestSourceHandler_Fetch(t *testing.T) {
	t.Run("rejects non-url source", func(t *testing.T) {
		store := newStubSourceCRUDStore()
		store.items[1] = models.Source{Id: 1, Type: "text", Status: "ready"}
		h := makeSourceHandler(store)

		req := httptest.NewRequest(http.MethodPost, "/sources/1/fetch", nil)
		req.SetPathValue("id", "1")
		resp := httptest.NewRecorder()
		h.Fetch(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", resp.Code, http.StatusBadRequest)
		}
	})

	t.Run("sets pending and clears content for url source", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "<html><body><p>Test content</p></body></html>")
		}))
		defer ts.Close()

		store := newStubSourceCRUDStore()
		store.items[1] = models.Source{Id: 1, Type: "url", Raw: ts.URL, Status: "ready", Content: "old content"}
		h := makeSourceHandler(store)

		req := httptest.NewRequest(http.MethodPost, "/sources/1/fetch", nil)
		req.SetPathValue("id", "1")
		resp := httptest.NewRecorder()
		h.Fetch(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
		}

		updated := store.getUpdated(1)
		if updated.Status != "pending" {
			t.Errorf("status = %q, want %q", updated.Status, "pending")
		}
		if updated.Content != "" {
			t.Errorf("content = %q, want empty (cleared)", updated.Content)
		}
	})
}
