package store

import (
	"maps"
	"sync"
)

type InMemoryStorage struct {
	mutex sync.RWMutex
	data  map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		mutex: sync.RWMutex{},
		data:  make(map[string]string),
	}
}

func (s *InMemoryStorage) Get(key string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	val, ok := s.data[key]
	if ok {
		return val, nil
	}
	return "", ErrKeyNotFound
}

func (s *InMemoryStorage) Set(key string, val string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[key] = val
	return nil
}

func (s *InMemoryStorage) Delete(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.data, key)
	return nil
}

func (s *InMemoryStorage) GetSnapshot() map[string]string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return maps.Clone(s.data)
}

func (s *InMemoryStorage) ApplySnapshot(data map[string]string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data = data
}
