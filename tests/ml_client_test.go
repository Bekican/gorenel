package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bekican/gorenel/internal/ml"
	"go.uber.org/zap"
)

func TestMLClientTimeout(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// --- MOCK SERVER (GÖREV SENDE!) ---
	// Bu sunucu, kendisine gelen isteklere 5 saniye geç cevap vermeli.
	// Bizim Go client'ımızın timeout süresi 3 saniye (veya ctx ile 2s) olduğu için
	// client'ın "context deadline exceeded" hatası almasını bekliyoruz.

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sunucunun 5 saniye uyumasını sağlıyoruz
		time.Sleep(5 * time.Second)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "models": {}, "consensus": {"any_anomaly": false}}`))
	}))
	defer ts.Close()

	client := ml.NewClient(ts.URL, logger)

	// 2 saniyelik bir context oluşturuyoruz
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	_, err := client.PredictCompare(ctx, map[string]interface{}{"test": "data"})
	duration := time.Since(start)

	if err == nil {
		t.Fatal("Hata alınmalıydı (timeout), ama istek başarılı oldu!")
	}

	fmt.Printf("İstek durduruldu. Süre: %v, Hata: %v\n", duration, err)

	if duration > 3*time.Second {
		t.Errorf("Timeout mekanizması yavaş kaldı: %v", duration)
	}
}
