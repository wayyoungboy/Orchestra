package a2a

import (
	"sync"
)

// AgentRegistry maps workspace:member keys to A2A agent endpoints.
type AgentRegistry struct {
	mu     sync.RWMutex
	agents map[string]*AgentEntry // "workspaceID:memberID" -> entry
}

// AgentEntry holds metadata for an A2A-capable agent.
type AgentEntry struct {
	MemberID  string
	AgentURL  string
	AuthType  string
	AuthToken string
}

// NewAgentRegistry creates an empty agent registry.
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]*AgentEntry),
	}
}

// Register adds or updates an agent entry.
func (r *AgentRegistry) Register(key string, entry *AgentEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[key] = entry
}

// Get retrieves an agent entry by key.
func (r *AgentRegistry) Get(key string) *AgentEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.agents[key]
}

// Unregister removes an agent entry.
func (r *AgentRegistry) Unregister(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.agents, key)
}
