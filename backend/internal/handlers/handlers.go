package handlers

import (
	"backend/internal/models"
	"encoding/json"
	"net/http"
	"strconv"
)

type Store[T any] interface {
	GetAll() ([]T, error)
	Get(id int) (T, error)
	Create(item T) (T, error)
	Delete(id int) error
}

type CrudHandler[T models.Identifiable] struct {
	store Store[T]
}

func NewCrudHandler[T models.Identifiable](s Store[T]) *CrudHandler[T] {
	return &CrudHandler[T]{store: s}
}

func (h *CrudHandler[T]) GetAll(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.GetAll()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *CrudHandler[T]) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	item, err := h.store.Get(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (h *CrudHandler[T]) Create(w http.ResponseWriter, r *http.Request) {
	var item T
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	created, err := h.store.Create(item)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *CrudHandler[T]) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.store.Delete(id); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
