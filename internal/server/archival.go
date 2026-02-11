package server

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
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

	done   chan struct{}
	logger *zap.Logger
}

// yeni data arşivi oluşturucaz
func NewDataArchiver(archiveDir string, rotateInterval time.Duration, retentionDays int) (*DataArchiver, error) {
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return nil, fmt.Errorf("archive dir oluşturulamadı:%w", err)
	}
	l, _ := zap.NewProduction()
	da := &DataArchiver{
		archiveDir:     archiveDir,
		rotateInterval: rotateInterval,
		retentionDays:  retentionDays,
		done:           make(chan struct{}),
		logger:         l,
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
		da.logger.Info("Archive dosyası kapatıldı", zap.Int64("events", da.eventsInFile))
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

	da.logger.Info("Yeni archive dosyası", zap.String("filename", filename))

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
				da.logger.Error("File rotation hatası", zap.Error(err))
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
		da.logger.Error("Archive dosyaları okunamadı", zap.Error(err))
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
				da.logger.Warn("Dosya silinemedi", zap.String("file", file))
			} else {
				deletedCount++
			}
		}
	}
	if deletedCount > 0 {
		da.logger.Info("Eski archive dosyaları silindi", zap.Int("count", deletedCount), zap.Int("retention_days", da.retentionDays))
	}
}

// archiver'ı kapatıyoruz
func (da *DataArchiver) Close() error {
	close(da.done)

	da.fileMu.Lock()
	defer da.fileMu.Unlock()

	if da.gzipWriter != nil {
		da.gzipWriter.Close()
	}

	if da.currentFile != nil {
		da.currentFile.Close()
	}

	return nil
}

// istatistikler
func (da *DataArchiver) Stats() map[string]interface{} {
	da.mu.RLock()
	defer da.mu.RUnlock()

	da.fileMu.Lock()
	currentEvents := da.eventsInFile
	da.fileMu.Unlock()

	return map[string]interface{}{
		"total_archived":    da.TotalArchived,
		"total_files":       da.TotalFiles,
		"current_file_size": currentEvents,
		"rotate_interval":   da.rotateInterval.String(),
		"retention_days":    da.retentionDays,
	}
}

// AWS S3'e nasıl export edilir onu anlatıyoruz - sorumluluk almıyoruz
func (da *DataArchiver) ExportToS3(bucketName string) string {
	return fmt.Sprintf(`
# S3'e upload için (AWS CLI):
aws s3 sync %s s3://%s/gorenel-archives/ \
  --storage-class GLACIER \
  --exclude "*" \
  --include "archive_*.jsonl.gz"

# Glacier'a direct upload (cost-effective):
aws s3 cp %s s3://%s/gorenel-archives/ \
  --recursive \
  --storage-class DEEP_ARCHIVE
`, da.archiveDir, bucketName, da.archiveDir, bucketName)
}

// archiveden eventleri okuyoruz
func RestoreFromArchive(archiveFile string) ([]*RequestEvent, error) {
	file, err := os.Open(archiveFile)
	if err != nil {
		return nil, fmt.Errorf("archive açılamadı : %w", err)
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("gzip okunamadı : %w", err)
	}
	defer gzipReader.Close()

	decoder := json.NewDecoder(gzipReader)
	events := make([]*RequestEvent, 0)

	for {
		var event RequestEvent
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("event decode edilemedi: %w", err)
		}
		events = append(events, &event)
	}
	return events, nil
}
