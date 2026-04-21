package provider

import (
	"sync"
)

// Registry manages all available providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[ProviderName]AgentProvider
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[ProviderName]AgentProvider),
	}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p AgentProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

// Get returns a provider by name, or nil if not found.
func (r *Registry) Get(name ProviderName) AgentProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[name]
}

// List returns all registered providers.
func (r *Registry) List() map[ProviderName]AgentProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[ProviderName]AgentProvider, len(r.providers))
	for k, v := range r.providers {
		result[k] = v
	}
	return result
}

// Installed returns only the providers that are currently installed on the system.
func (r *Registry) Installed() []AgentProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var installed []AgentProvider
	for _, p := range r.providers {
		if p.IsInstalled() {
			installed = append(installed, p)
		}
	}
	return installed
}
