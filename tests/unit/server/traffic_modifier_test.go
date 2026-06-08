package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bekican/gorenel/internal/server"
	"github.com/stretchr/testify/assert"
)

// ===== HEADER MODIFICATION =====

func TestTrafficModifier_AddHeaders(t *testing.T) {
	mod := server.NewTrafficModifier()
	mod.AddRule(server.ModificationRule{
		ID:          "add-auth",
		PathPattern: "/api/*",
		AddHeaders:  map[string]string{"X-Custom-Auth": "token-123"},
	})

	req := httptest.NewRequest("GET", "http://example.com/api/users", nil)
	mod.Apply(req)

	assert.Equal(t, "token-123", req.Header.Get("X-Custom-Auth"), "Should inject custom header")
}

func TestTrafficModifier_RemoveHeaders(t *testing.T) {
	mod := server.NewTrafficModifier()
	mod.AddRule(server.ModificationRule{
		ID:            "strip-cookie",
		PathPattern:   "/api/*",
		RemoveHeaders: []string{"Cookie", "X-Debug"},
	})

	req := httptest.NewRequest("GET", "http://example.com/api/data", nil)
	req.Header.Set("Cookie", "session=abc")
	req.Header.Set("X-Debug", "true")
	req.Header.Set("Accept", "application/json")

	mod.Apply(req)

	assert.Empty(t, req.Header.Get("Cookie"), "Cookie header should be removed")
	assert.Empty(t, req.Header.Get("X-Debug"), "X-Debug header should be removed")
	assert.Equal(t, "application/json", req.Header.Get("Accept"), "Accept should be untouched")
}

// ===== PATH REPLACEMENT =====

func TestTrafficModifier_ReplacePath(t *testing.T) {
	mod := server.NewTrafficModifier()
	mod.AddRule(server.ModificationRule{
		ID:          "redirect-path",
		PathPattern: "/old-api",
		ReplacePath: "/v2/api",
	})

	req := httptest.NewRequest("GET", "http://example.com/old-api", nil)
	mod.Apply(req)

	assert.Equal(t, "/v2/api", req.URL.Path, "Path should be replaced")
}

// ===== GLOB PATTERN MATCHING =====

func TestTrafficModifier_GlobMatchWildcard(t *testing.T) {
	mod := server.NewTrafficModifier()
	mod.AddRule(server.ModificationRule{
		ID:          "glob-test",
		PathPattern: "/api/*",
		AddHeaders:  map[string]string{"X-Matched": "true"},
	})

	tests := []struct {
		path    string
		matched bool
	}{
		{"/api/users", true},
		{"/api/products/123", true},
		{"/api/", true},
		{"/other/path", false},
		{"/", false},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", "http://example.com"+tt.path, nil)
		mod.Apply(req)

		if tt.matched {
			assert.Equal(t, "true", req.Header.Get("X-Matched"), "Path %s should match", tt.path)
		} else {
			assert.Empty(t, req.Header.Get("X-Matched"), "Path %s should NOT match", tt.path)
		}
	}
}

func TestTrafficModifier_ExactMatch(t *testing.T) {
	mod := server.NewTrafficModifier()
	mod.AddRule(server.ModificationRule{
		ID:          "exact",
		PathPattern: "/health",
		AddHeaders:  map[string]string{"X-Health": "true"},
	})

	req1 := httptest.NewRequest("GET", "http://example.com/health", nil)
	mod.Apply(req1)
	assert.Equal(t, "true", req1.Header.Get("X-Health"), "Exact match should work")

	req2 := httptest.NewRequest("GET", "http://example.com/health/detail", nil)
	mod.Apply(req2)
	assert.Empty(t, req2.Header.Get("X-Health"), "Should not match sub-paths without wildcard")
}

// ===== RULE CRUD =====

func TestTrafficModifier_AddRemoveRules(t *testing.T) {
	mod := server.NewTrafficModifier()

	mod.AddRule(server.ModificationRule{ID: "rule-1", PathPattern: "/a"})
	mod.AddRule(server.ModificationRule{ID: "rule-2", PathPattern: "/b"})
	mod.AddRule(server.ModificationRule{ID: "rule-3", PathPattern: "/c"})

	rules := mod.GetRules()
	assert.Len(t, rules, 3, "Should have 3 rules")

	mod.RemoveRule("rule-2")
	rules = mod.GetRules()
	assert.Len(t, rules, 2, "Should have 2 rules after removal")

	for _, r := range rules {
		assert.NotEqual(t, "rule-2", r.ID, "rule-2 should be removed")
	}
}

func TestTrafficModifier_RemoveNonExistent(t *testing.T) {
	mod := server.NewTrafficModifier()
	mod.AddRule(server.ModificationRule{ID: "only-rule", PathPattern: "/"})

	// Should not panic
	mod.RemoveRule("ghost-rule")
	assert.Len(t, mod.GetRules(), 1, "Original rule should still be there")
}

// ===== NO-MATCH PASS-THROUGH =====

func TestTrafficModifier_NoMatchPassThrough(t *testing.T) {
	mod := server.NewTrafficModifier()
	mod.AddRule(server.ModificationRule{
		ID:          "api-only",
		PathPattern: "/api/*",
		AddHeaders:  map[string]string{"X-Modified": "true"},
	})

	req := httptest.NewRequest("GET", "http://example.com/static/image.png", nil)
	req.Header.Set("Original", "untouched")
	mod.Apply(req)

	assert.Empty(t, req.Header.Get("X-Modified"), "Non-matching request should not be modified")
	assert.Equal(t, "untouched", req.Header.Get("Original"), "Original headers should remain")
}

// ===== MULTIPLE RULES =====

func TestTrafficModifier_MultipleRulesApplied(t *testing.T) {
	mod := server.NewTrafficModifier()
	mod.AddRule(server.ModificationRule{
		ID:          "cors",
		PathPattern: "/api/*",
		AddHeaders:  map[string]string{"Access-Control-Allow-Origin": "*"},
	})
	mod.AddRule(server.ModificationRule{
		ID:          "auth",
		PathPattern: "/api/*",
		AddHeaders:  map[string]string{"X-Auth-Required": "true"},
	})

	req := httptest.NewRequest("GET", "http://example.com/api/users", nil)
	mod.Apply(req)

	assert.Equal(t, "*", req.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", req.Header.Get("X-Auth-Required"))
}

// ===== GetRules RETURNS COPY =====

func TestTrafficModifier_GetRulesReturnsCopy(t *testing.T) {
	mod := server.NewTrafficModifier()
	mod.AddRule(server.ModificationRule{ID: "original", PathPattern: "/"})

	rules := mod.GetRules()
	rules[0].ID = "mutated"

	originalRules := mod.GetRules()
	assert.Equal(t, "original", originalRules[0].ID, "Mutation should not affect internal state")
}

// ===== BENCHMARK =====

func BenchmarkTrafficModifier_Apply(b *testing.B) {
	mod := server.NewTrafficModifier()
	for i := 0; i < 10; i++ {
		mod.AddRule(server.ModificationRule{
			ID:          http.StatusText(i),
			PathPattern: "/api/*",
			AddHeaders:  map[string]string{"X-Rule": "applied"},
		})
	}

	req := httptest.NewRequest("GET", "http://example.com/api/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod.Apply(req)
	}
}
