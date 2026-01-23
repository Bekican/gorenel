package server

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

func (da *DataArchiver) Name() string {
	return "DataArchiver"
}

func (da *DataArchiver) rotateFile() error {
	da.fileMu.Lock()
	defer da.fileMu.Unlock()

	return da.rotateFileLocked()
}

// rotateFileLocked
func (da *DataArchiver) rotateFileLocked() error {
	if da.gzipWriter != nil {
		da.gzipWriter.Close()
	}
	if da.currentFile != nil {
		da.currentFile.Close()
		log.Printf("archive dosyası kapatıldı : %d events", da.eventsInFile)
	}

	//yeni dosya oluşturuyoruz
	filename := fmt.Sprintf("archive_%s.jsonl.gz", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(da.archiveDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("archive dosyası oluşturulamadı: %w", err)
	}
	gzipWriter := gzip.NewWriter(file)

	da.currentFile = file
	da.gzipWriter = gzipWriter
	da.encoder = json.NewEncoder(gzipWriter)
	da.eventsInFile = 0
	da.fileStartTime = time.Now()

	da.mu.Lock()
	da.TotalFiles++
	da.mu.Unlock()

	log.Printf("Yeni archive dosyası: %s", filename)

	return nil
}

// periyodik rotation ve cleanup
func (da *DataArchiver) periodicMaintenance() {
	//rotateticker
	rotateTicker := time.NewTicker(da.rotateInterval)
	defer rotateTicker.Stop()
	//cleanticker
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-rotateTicker.C:
			if err := da.rotateFile(); err != nil {
				log.Printf("File rotation hatası: %v", err)
			}
		case <-cleanupTicker.C:
			da.cleanup()
		case <-da.done:
			return
		}
	}
}

// eski archive dosyalarını siliyoruz
func (da *DataArchiver) cleanup() {
	cutoff := time.Now().AddDate(0, 0, -da.retentionDays)

	files, err := filepath.Glob(filepath.Join(da.archiveDir, "archive_*.jsonl.gz"))
	if err != nil {
		log.Printf("Archive dosyaları okunamadı : %v", err)
		return
	}

	deletedCount := 0
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		//retention süresine bakıyoruz geçmişse siliyoruz
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(file); err != nil {
				log.Printf("Dosya silinemedi: %s", file)
			} else {
				deletedCount++
			}
		}
	}
	if deletedCount > 0 {
		log.Printf("%d eski archive dosyaları silindi(retention:%d gün)", deletedCount, da.retentionDays)
	}
}

//archiver'ı kapatıyoruz
