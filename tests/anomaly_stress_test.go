package tests

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Bekican/gorenel/internal/server"
	"github.com/google/uuid"
)

// TestAnomalyStoreConcurrency verifies thread safety of Add and GetRecent
func TestAnomalyStoreConcurrency(t *testing.T) {
	const (
		numWriters = 50
		numReaders = 50
		iterations = 100
		maxRecords = 100
	)

	store := server.NewAnomalyStore(maxRecords)
	var wg sync.WaitGroup

	// --- WRITER LOGIC ---
	wg.Add(numWriters)
	for i := 0; i < numWriters; i++ {
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				store.Add(server.AnomalyRecord{
					ID:        uuid.New().String(),
					Timestamp: time.Now(),
					Subdomain: fmt.Sprintf("writer-%d", writerID),
					Method:    "GET",
					Path:      "/stress-test",
				})
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	// --- READER LOGIC (Implemented by User, Fixed by Assistant) ---
	wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go func(readerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// GetRecent requires a limit argument
				recent := store.GetRecent(10)

				if len(recent) > maxRecords {
					t.Errorf("GetRecent returned %d records > maxRecords %d", len(recent), maxRecords)
				}

				for _, r := range recent {
					if r.ID == "" {
						t.Errorf("Empty ID detected at reader %d", readerID)
					}
					if r.Subdomain == "" {
						t.Errorf("Empty subdomain detected at reader %d", readerID)
					}
				}
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	finalCount := len(store.GetRecent(maxRecords + 50))
	fmt.Printf("\n[DEBUG] Stress test completed. Final records in store: %d\n", finalCount)

	if finalCount > maxRecords {
		t.Errorf("Ring buffer overflow: %d > %d", finalCount, maxRecords)
	}
}
