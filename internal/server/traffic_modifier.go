package server

import (
	"net/http"
	"strings"
	"sync"
)

// ModificationRule represents a single rule to modify a request
type ModificationRule struct {
	ID            string            `json:"id"`
	PathPattern   string            `json:"path_pattern"` // e.g., "/api/*"
	AddHeaders    map[string]string `json:"add_headers,omitempty"`
	RemoveHeaders []string          `json:"remove_headers,omitempty"`
	ReplacePath   string            `json:"replace_path,omitempty"`
}

// TrafficModifier manages a set of rules and applies them to incoming requests
type TrafficModifier struct {
	mu    sync.RWMutex
	rules []ModificationRule
}

func NewTrafficModifier() *TrafficModifier {
	return &TrafficModifier{
		rules: make([]ModificationRule, 0),
	}
}

// Apply checks if the request matches any rules and modifies it in-place
func (m *TrafficModifier) Apply(r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, rule := range m.rules {
		if m.matches(r.URL.Path, rule.PathPattern) {
			// Apply Header Modifications
			for k, v := range rule.AddHeaders {
				r.Header.Set(k, v)
			}
			for _, k := range rule.RemoveHeaders {
				r.Header.Del(k)
			}

			// Apply Path Replacement if specified
			if rule.ReplacePath != "" {
				r.URL.Path = rule.ReplacePath
			}
		}
	}
}

// matches performs a simple glob-like matching for paths
func (m *TrafficModifier) matches(path, pattern string) bool {
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(path, strings.TrimSuffix(pattern, "*"))
	}
	return path == pattern
}

// AddRule adds a new rule to the modifier
func (m *TrafficModifier) AddRule(rule ModificationRule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rules = append(m.rules, rule)
}

// RemoveRule deletes a rule by its ID
func (m *TrafficModifier) RemoveRule(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, rule := range m.rules {
		if rule.ID == id {
			m.rules = append(m.rules[:i], m.rules[i+1:]...)
			return
		}
	}
}

// GetRules returns a copy of the current rules
func (m *TrafficModifier) GetRules() []ModificationRule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rules := make([]ModificationRule, len(m.rules))
	copy(rules, m.rules)
	return rules
}
