package server

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type DataArchiver struct {
	archiveDir     string
	rotateInterval time.Duration
	retentionDays  int

	//current arşiv
	currentFile   *os.File
	gzipWriter    *gzip.Writer
	encoder       *json.Encoder
	eventsInFile  int64
	fileStartTime time.Time
	fileMu        sync.Mutex

	//metricler
	TotalArchived int64
	TotalFiles    int64
	mu            sync.RWMutex

	done chan struct{}
}

// yeni data arşivi oluşturucaz
func NewDataArchiver(archiveDir string, rotateInterval time.Duration, retentionDays int) (*DataArchiver, error) {
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return nil, fmt.Errorf("archive dir oluşturulamadı:%w", err)
	}
	da := &DataArchiver{
		archiveDir:     archiveDir,
		rotateInterval: rotateInterval,
		retentionDays:  retentionDays,
		done:           make(chan struct{}),
	}
	//ilk dosyayı aç
	if err := da.rotateFile(); err != nil {
		return nil, err
	}
	go da.periodicMaintenance()

	return da, nil
}

// eventconsumer interface
func (da *DataArchiver) Consume(event *RequestEvent) error {
	da.fileMu.Lock()
	defer da.fileMu.Unlock()
	//logların eski dosyaya yazılmaması için / rotateFileLocked -> zaten işlem yapıysoun
	if time.Since(da.fileStartTime) > da.rotateInterval {
		if err := da.rotateFileLocked(); err != nil {
			return err
		}
	}

	//eventi yazdırıyoruz.
	if err := da.encoder.Encode(event); err != nil {
		return fmt.Errorf("event yazılamadı :%w", err)
	}

	da.eventsInFile++
	da.mu.Lock()
	da.TotalArchived++
	da.mu.Unlock()

	return nil
}
