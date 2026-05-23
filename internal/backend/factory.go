package backend

import (
	"fmt"

	"github.com/RubnMC/rubipaper/internal/domain"
)

// NewBackend returns a Backend for the given name, or an error if the name is
// unknown or the backend is not available on this system.
func NewBackend(name string) (domain.Backend, error) {
	backends := map[string]domain.Backend{
		"swaybg": NewSwaybgBackend(),
	}

	b, ok := backends[name]
	if !ok {
		available := make([]string, 0, len(backends))
		for k := range backends {
			available = append(available, k)
		}
		return nil, fmt.Errorf("unknown backend %q, available: %v", name, available)
	}

	if !b.IsAvaliable() {
		return nil, fmt.Errorf("backend %q is not available on this system", name)
	}

	return b, nil
}
