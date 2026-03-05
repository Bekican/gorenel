package server_test

import (
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/Bekican/gorenel/internal/server"
	"github.com/Bekican/gorenel/tests"
	"github.com/hashicorp/yamux"
	"github.com/stretchr/testify/assert"
)

// ===== LIFECYCLE TESTS =====

func TestTunnelManager_RegisterAndGet(t *testing.T) {
	helper := tests.NewTestHelper(t)
	tm := server.NewTunnelManager()

	serverSession, clientSession, err := helper.CreateYamuxPair()
	helper.RequireNoError(err, "Failed to create yamux pair")
	defer serverSession.Close()
	defer clientSession.Close()

	tm.RegisterTunnel("test-sub", serverSession, "", 3000, "http://test-sub.gorenel.net:8080")

	session, exists := tm.GetTunnel("test-sub")
	assert.True(t, exists, "Tunnel should exist after registration")
	assert.NotNil(t, session, "Session should not be nil")
	assert.Equal(t, 1, tm.Count(), "Should have 1 active tunnel")
}

func TestTunnelManager_RemoveTunnel(t *testing.T) {
	helper := tests.NewTestHelper(t)
	tm := server.NewTunnelManager()

	serverSession, clientSession, err := helper.CreateYamuxPair()
	helper.RequireNoError(err, "Failed to create yamux pair")
	defer serverSession.Close()
	defer clientSession.Close()

	tm.RegisterTunnel("remove-me", serverSession, "", 3000, "http://remove-me.gorenel.net:8080")
	assert.Equal(t, 1, tm.Count())

	tm.RemoveTunnel("remove-me")

	_, exists := tm.GetTunnel("remove-me")
	assert.False(t, exists, "Tunnel should not exist after removal")
	assert.Equal(t, 0, tm.Count(), "Should have 0 tunnels after removal")
}

func TestTunnelManager_GetNonExistent(t *testing.T) {
	tm := server.NewTunnelManager()

	_, exists := tm.GetTunnel("does-not-exist")
	assert.False(t, exists, "Non-existent tunnel should return false")
}

// ===== CUSTOM DOMAIN TESTS =====

func TestTunnelManager_CustomDomain(t *testing.T) {
	helper := tests.NewTestHelper(t)
	tm := server.NewTunnelManager()

	serverSession, clientSession, err := helper.CreateYamuxPair()
	helper.RequireNoError(err, "Failed to create yamux pair")
	defer serverSession.Close()
	defer clientSession.Close()

	tm.RegisterTunnel("my-sub", serverSession, "api.bekircan.com", 3000, "http://my-sub.gorenel.net:8080")

	_, existsBySub := tm.GetTunnel("my-sub")
	assert.True(t, existsBySub, "Should find by subdomain")

	_, existsByDomain := tm.GetTunnel("api.bekircan.com")
	assert.True(t, existsByDomain, "Should find by custom domain")

	tm.RemoveTunnel("my-sub")
	_, existsByDomain = tm.GetTunnel("api.bekircan.com")
	assert.False(t, existsByDomain, "Custom domain should be cleared after tunnel removal")
}

// ===== STATS TESTS =====

func TestTunnelManager_UpdateStats(t *testing.T) {
	helper := tests.NewTestHelper(t)
	tm := server.NewTunnelManager()

	serverSession, clientSession, err := helper.CreateYamuxPair()
	helper.RequireNoError(err, "Failed to create yamux pair")
	defer serverSession.Close()
	defer clientSession.Close()

	tm.RegisterTunnel("stats-sub", serverSession, "", 3000, "http://stats-sub.gorenel.net:8080")

	tm.UpdateStats("stats-sub", 1024, 2048)
	tm.UpdateStats("stats-sub", 512, 256)

	tunnels := tm.GetTunnels()
	assert.Len(t, tunnels, 1)
	assert.Equal(t, int64(2), tunnels[0].RequestCount, "Should count 2 requests")
	assert.Equal(t, int64(1536), tunnels[0].Bandwidth.In, "Bytes in should be 1024+512")
	assert.Equal(t, int64(2304), tunnels[0].Bandwidth.Out, "Bytes out should be 2048+256")
}

func TestTunnelManager_UpdateStatsNonExistent(t *testing.T) {
	tm := server.NewTunnelManager()
	// Should not panic on non-existent tunnel
	tm.UpdateStats("ghost", 100, 200)
	assert.Equal(t, 0, tm.Count())
}

// ===== MULTIPLE TUNNELS =====

func TestTunnelManager_MultipleTunnels(t *testing.T) {
	helper := tests.NewTestHelper(t)
	tm := server.NewTunnelManager()

	for i := 0; i < 5; i++ {
		s, c, err := helper.CreateYamuxPair()
		helper.RequireNoError(err, "pair creation")
		defer s.Close()
		defer c.Close()
		tm.RegisterTunnel(fmt.Sprintf("sub-%d", i), s, "", 3000+i, "http://test.gorenel.net:8080")
	}

	assert.Equal(t, 5, tm.Count())
	tunnels := tm.GetTunnels()
	assert.Len(t, tunnels, 5)

	// Remove one and verify
	tm.RemoveTunnel("sub-2")
	assert.Equal(t, 4, tm.Count())
}

// ===== CONCURRENCY =====

func TestTunnelManager_ConcurrentRegisterRemove(t *testing.T) {
	helper := tests.NewTestHelper(t)
	tm := server.NewTunnelManager()
	numTunnels := 50

	var wg sync.WaitGroup

	wg.Add(numTunnels)
	for i := 0; i < numTunnels; i++ {
		go func(idx int) {
			defer wg.Done()
			s, c, err := helper.CreateYamuxPair()
			if err != nil {
				return
			}
			defer c.Close()
			sub := fmt.Sprintf("concurrent-%d", idx)
			tm.RegisterTunnel(sub, s, "", 3000+idx, fmt.Sprintf("http://%s.gorenel.net:8080", sub))
		}(i)
	}
	wg.Wait()
	assert.Equal(t, numTunnels, tm.Count(), "All tunnels should be registered")

	wg.Add(numTunnels)
	for i := 0; i < numTunnels; i++ {
		go func(idx int) {
			defer wg.Done()
			tm.RemoveTunnel(fmt.Sprintf("concurrent-%d", idx))
		}(i)
	}
	wg.Wait()
	assert.Equal(t, 0, tm.Count(), "All tunnels should be removed")
}

// ===== BENCHMARKS =====

func BenchmarkTunnelManager_RegisterTunnel(b *testing.B) {
	tm := server.NewTunnelManager()
	s, c := mustCreateYamuxPair(b)
	defer s.Close()
	defer c.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sub := fmt.Sprintf("bench-%d", i)
		tm.RegisterTunnel(sub, s, "", 3000, "http://bench.gorenel.net:8080")
	}
}

func BenchmarkTunnelManager_GetTunnel(b *testing.B) {
	tm := server.NewTunnelManager()
	s, c := mustCreateYamuxPair(b)
	defer s.Close()
	defer c.Close()

	tm.RegisterTunnel("bench-lookup", s, "", 3000, "http://bench.gorenel.net:8080")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.GetTunnel("bench-lookup")
	}
}

func mustCreateYamuxPair(b *testing.B) (*yamux.Session, *yamux.Session) {
	b.Helper()
	sConn, cConn := net.Pipe()
	s, err := yamux.Server(sConn, yamux.DefaultConfig())
	if err != nil {
		b.Fatal(err)
	}
	c, err := yamux.Client(cConn, yamux.DefaultConfig())
	if err != nil {
		b.Fatal(err)
	}
	return s, c
}
