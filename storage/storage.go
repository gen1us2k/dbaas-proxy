package storage

import (
	"sync"

	"k8s.io/client-go/rest"
)

type (
	Storage struct {
		mu    *sync.Mutex
		items map[string]*rest.Config
	}
)

func New() *Storage {
	return &Storage{
		mu:    &sync.Mutex{},
		items: make(map[string]*rest.Config),
	}
}

func (s *Storage) Add(name string, kubeconfig *rest.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[name] = kubeconfig
}
func (s *Storage) Get(name string) *rest.Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.items[name]
}

func (s *Storage) Delete(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, name)
}
