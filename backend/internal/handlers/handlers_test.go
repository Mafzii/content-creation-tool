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
	"testing"
)

type stubStore[T models.Identifiable] struct {
	items []T
	err   error
}

func (s *stubStore[T]) GetAll() ([]T, error) {
	return s.items, s.err
}

func (s *stubStore[T]) Get(id int) (T, error) {
	if s.err != nil {
		var zero T
		return zero, s.err
	}
	for _, item := range s.items {
		if item.GetId() == id {
			return item, nil
		}
	}
	var zero T
	return zero, errors.New("not found")
}

func (s *stubStore[T]) Create(item T) (T, error) {
	if s.err != nil {
		var zero T
		return zero, s.err
	}
	s.items = append(s.items, item)
	return item, nil
}

func (s *stubStore[T]) Delete(id int) error {
	if s.err != nil {
		return s.err
	}
	for i, item := range s.items {
		if item.GetId() == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}

func (s *stubStore[T]) Update(id int, item T) (T, error) {
	if s.err != nil {
		var zero T
		return zero, s.err
	}
	for i, existing := range s.items {
		if existing.GetId() == id {
			s.items[i] = item
			return item, nil
		}
	}
	var zero T
	return zero, errors.New("not found")
}

type crudIface interface {
	GetAll(http.ResponseWriter, *http.Request)
	Get(http.ResponseWriter, *http.Request)
	Create(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
}

func newTopicHandler(items []models.Topic, err error) crudIface {
	return handlers.NewCrudHandler[models.Topic](&stubStore[models.Topic]{items: items, err: err})
}


type entityTest struct {
	name        string
	makeHandler func(seedJSON string, err error) crudIface
	seedJSON    string
	newJSON     string
	entityID    string
	missingID   string
}

func unmarshalHandler[T models.Identifiable](maker func([]T, error) crudIface) func(string, error) crudIface {
	return func(seedJSON string, storeErr error) crudIface {
		if storeErr != nil {
			return maker(nil, storeErr)
		}
		var items []T
		json.Unmarshal([]byte(seedJSON), &items)
		return maker(items, nil)
	}
}

func TestCrudHandler(t *testing.T) {
	cases := []entityTest{
		{
			name:        "Topic",
			makeHandler: unmarshalHandler(newTopicHandler),
			seedJSON:    `[{"id":1,"name":"computers"},{"id":2,"name":"science"}]`,
			newJSON:     `{"id":3,"name":"mathematics"}`,
			entityID:    "1",
			missingID:   "99",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("GetAll returns items", func(t *testing.T) {
				h := tc.makeHandler(tc.seedJSON, nil)
				req, _ := http.NewRequest(http.MethodGet, "/", nil)
				resp := httptest.NewRecorder()
				h.GetAll(resp, req)
				if resp.Code != http.StatusOK {
					t.Errorf("got status %d, want %d", resp.Code, http.StatusOK)
				}
			})

			t.Run("GetAll returns 500 on error", func(t *testing.T) {
				h := tc.makeHandler("", errors.New("db down"))
				req, _ := http.NewRequest(http.MethodGet, "/", nil)
				resp := httptest.NewRecorder()
				h.GetAll(resp, req)
				if resp.Code != http.StatusInternalServerError {
					t.Errorf("got status %d, want %d", resp.Code, http.StatusInternalServerError)
				}
			})

			t.Run("Get returns item by id", func(t *testing.T) {
				h := tc.makeHandler(tc.seedJSON, nil)
				req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", tc.entityID), nil)
				req.SetPathValue("id", tc.entityID)
				resp := httptest.NewRecorder()
				h.Get(resp, req)
				if resp.Code != http.StatusOK {
					t.Errorf("got status %d, want %d", resp.Code, http.StatusOK)
				}
			})

			t.Run("Get returns 404 for missing item", func(t *testing.T) {
				h := tc.makeHandler(tc.seedJSON, nil)
				req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", tc.missingID), nil)
				req.SetPathValue("id", tc.missingID)
				resp := httptest.NewRecorder()
				h.Get(resp, req)
				if resp.Code != http.StatusNotFound {
					t.Errorf("got status %d, want %d", resp.Code, http.StatusNotFound)
				}
			})

			t.Run("Create returns 201", func(t *testing.T) {
				h := tc.makeHandler("[]", nil)
				req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tc.newJSON)))
				resp := httptest.NewRecorder()
				h.Create(resp, req)
				if resp.Code != http.StatusCreated {
					t.Errorf("got status %d, want %d", resp.Code, http.StatusCreated)
				}
			})

			t.Run("Create returns 500 on error", func(t *testing.T) {
				h := tc.makeHandler("", errors.New("db down"))
				req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tc.newJSON)))
				resp := httptest.NewRecorder()
				h.Create(resp, req)
				if resp.Code != http.StatusInternalServerError {
					t.Errorf("got status %d, want %d", resp.Code, http.StatusInternalServerError)
				}
			})

			t.Run("Delete returns 200", func(t *testing.T) {
				h := tc.makeHandler(tc.seedJSON, nil)
				req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/%s", tc.entityID), nil)
				req.SetPathValue("id", tc.entityID)
				resp := httptest.NewRecorder()
				h.Delete(resp, req)
				if resp.Code != http.StatusOK {
					t.Errorf("got status %d, want %d", resp.Code, http.StatusOK)
				}
			})

			t.Run("Delete returns 404 for missing item", func(t *testing.T) {
				h := tc.makeHandler(tc.seedJSON, nil)
				req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/%s", tc.missingID), nil)
				req.SetPathValue("id", tc.missingID)
				resp := httptest.NewRecorder()
				h.Delete(resp, req)
				if resp.Code != http.StatusNotFound {
					t.Errorf("got status %d, want %d", resp.Code, http.StatusNotFound)
				}
			})

			t.Run("Update returns 200", func(t *testing.T) {
				h := tc.makeHandler(tc.seedJSON, nil)
				req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/%s", tc.entityID), bytes.NewReader([]byte(tc.newJSON)))
				req.SetPathValue("id", tc.entityID)
				resp := httptest.NewRecorder()
				h.Update(resp, req)
				if resp.Code != http.StatusOK {
					t.Errorf("got status %d, want %d", resp.Code, http.StatusOK)
				}
			})

			t.Run("Update returns 500 on error", func(t *testing.T) {
				h := tc.makeHandler("", errors.New("db down"))
				req, _ := http.NewRequest(http.MethodPut, "/1", bytes.NewReader([]byte(tc.newJSON)))
				req.SetPathValue("id", "1")
				resp := httptest.NewRecorder()
				h.Update(resp, req)
				if resp.Code != http.StatusInternalServerError {
					t.Errorf("got status %d, want %d", resp.Code, http.StatusInternalServerError)
				}
			})

			t.Run("Update returns 400 for bad id", func(t *testing.T) {
				h := tc.makeHandler(tc.seedJSON, nil)
				req, _ := http.NewRequest(http.MethodPut, "/abc", bytes.NewReader([]byte(tc.newJSON)))
				req.SetPathValue("id", "abc")
				resp := httptest.NewRecorder()
				h.Update(resp, req)
				if resp.Code != http.StatusBadRequest {
					t.Errorf("got status %d, want %d", resp.Code, http.StatusBadRequest)
				}
			})

			t.Run("Update returns 400 for bad JSON", func(t *testing.T) {
				h := tc.makeHandler(tc.seedJSON, nil)
				req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/%s", tc.entityID), bytes.NewReader([]byte("not json")))
				req.SetPathValue("id", tc.entityID)
				resp := httptest.NewRecorder()
				h.Update(resp, req)
				if resp.Code != http.StatusBadRequest {
					t.Errorf("got status %d, want %d", resp.Code, http.StatusBadRequest)
				}
			})
		})
	}
}
