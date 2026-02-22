package store

import (
	"backend/internal/models"
	"fmt"
)

type InMemoryStore[T models.Identifiable] struct {
	items []T
}

func NewInMemoryStore[T models.Identifiable](items []T) *InMemoryStore[T] {
	return &InMemoryStore[T]{items: items}
}

func (s *InMemoryStore[T]) GetAll() ([]T, error) {
	return s.items, nil
}

func (s *InMemoryStore[T]) Get(id int) (T, error) {
	for _, item := range s.items {
		if item.GetId() == id {
			return item, nil
		}
	}
	var zero T
	return zero, fmt.Errorf("item %d not found", id)
}

func (s *InMemoryStore[T]) Create(item T) (T, error) {
	s.items = append(s.items, item)
	return item, nil
}

func (s *InMemoryStore[T]) Delete(id int) error {
	for i, item := range s.items {
		if item.GetId() == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("item %d not found", id)
}
