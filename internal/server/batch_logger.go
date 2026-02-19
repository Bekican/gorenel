package server

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Bekican/gorenel/internal/analytics"
	"go.uber.org/zap"
)

type BatchLogger struct {
	batchSize     int
	flushInterval time.Duration

	buffer []*RequestEvent
	bufMu  sync.Mutex

	outputDir   string
	currentFile *os.File
	gzipWriter  *gzip.Writer

	chRepo *analytics.ClickHouseRepo

	TotalBatches int64
	TotalEvents  int64
	mu           sync.RWMutex

	done   chan struct{}
	logger *zap.Logger
}

// yeni batch logger oluşturuyoruz
func NewBatchLogger(outputDir string, batchSize int, flushInterval time.Duration, chRepo *analytics.ClickHouseRepo) (*BatchLogger, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("outputdir oluşturulamadı:%w", err)
	}

	l, _ := zap.NewProduction()
	bl := &BatchLogger{
		batchSize:     batchSize,
		flushInterval: flushInterval,
		buffer:        make([]*RequestEvent, 0, batchSize),
		outputDir:     outputDir,
		chRepo:        chRepo,
		done:          make(chan struct{}),
		logger:        l,
	}

	go bl.periodicFlush()

	return bl, nil

}

// consumer
func (bl *BatchLogger) Consume(event *RequestEvent) error {
	bl.bufMu.Lock()
	defer bl.bufMu.Unlock()

	bl.buffer = append(bl.buffer, event)

	if len(bl.buffer) > bl.batchSize {
		return bl.flushLocked()
	}
	return nil
}

// consumer adı
func (bl *BatchLogger) Name() string {
	return "BatchLogger"
}

// periodicFlush -- periyodik olarak flushInterval -> flush ediyoruz
func (bl *BatchLogger) periodicFlush() {
	ticker := time.NewTicker(bl.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bl.bufMu.Lock()
			if len(bl.buffer) > 0 {
				bl.flushLocked()
			}
			bl.bufMu.Unlock()

		case <-bl.done:
			return
		}
	}
}

// buffer'ı dosyaya yazma -- flushLocked
func (bl *BatchLogger) flushLocked() error {
	if len(bl.buffer) == 0 {
		return nil
	}

	// ClickHouse'a gönder
	if bl.chRepo != nil {
		go func(events []*RequestEvent) {
			if err := bl.chRepo.BatchInsert(events); err != nil {
				bl.logger.Error("ClickHouse insert hatası", zap.Error(err))
			}
		}(append([]*RequestEvent{}, bl.buffer...)) // Copy buffer for async insert
	}

	// Dosyaya da yaz (Backup)
	filename := fmt.Sprintf("events_%s.jsonl.gz", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(bl.outputDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("dosya oluşturulamadı:%w", err)
	}

	gzipWriter := gzip.NewWriter(file)
	encoder := json.NewEncoder(gzipWriter)

	for _, event := range bl.buffer {
		if err := encoder.Encode(event); err != nil {
			bl.logger.Error("Event yazılamadı", zap.Error(err))
			continue
		}
	}

	gzipWriter.Close()
	file.Close()

	bl.mu.Lock()
	bl.TotalBatches++
	bl.TotalEvents += int64(len(bl.buffer))
	bl.mu.Unlock()

	bl.logger.Info("Batch yazıldı", zap.String("filename", filename), zap.Int("events", len(bl.buffer)))

	// buffer1ı temizle
	bl.buffer = bl.buffer[:0]

	return nil

}

//batch logger'ı kapatıcaz

func (bl *BatchLogger) Close() error {
	close(bl.done)

	bl.bufMu.Unlock()
	defer bl.bufMu.Unlock()

	return bl.flushLocked()
}

// istatistikleri gösteriyoruz
func (bl *BatchLogger) Stats() map[string]interface{} {
	bl.mu.RUnlock()
	defer bl.mu.RUnlock()

	bl.bufMu.Lock()
	bufferSize := len(bl.buffer)
	bl.bufMu.Unlock()

	return map[string]interface{}{
		"total_batches":  bl.TotalBatches,
		"total_events":   bl.TotalEvents,
		"current_buffer": bufferSize,
		"batch_size":     bl.batchSize,
		"flush_interval": bl.flushInterval.String(),
	}
}

// clickhouse'a export
func (bl *BatchLogger) ExportToClickHouse(tableName string) string {
	return fmt.Sprintf(`
		clickhouse-client --query="
	INSERT INTO %s
	SELECT * FROM file('events_*.jsonl.gz',JSONEachRow)
	"
	`, tableName)
}

// elasticSearche bulk import
func (bl *BatchLogger) ExportToElasticsearch(indexname string) string {
	return fmt.Sprintf(`
	for file in events_*.jsonl.gz; do
	gunzip -c $file | curl -X POST "localhost:9200/%s/_bulk"\
	-H "Content-Type : application/x-ndjson"\
	--data-binary @-
	done
	`, indexname)
}
