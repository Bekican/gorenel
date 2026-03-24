package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	baseUrl    string
	httpClient *http.Client
	logger     *zap.Logger
}

type PredictionRequest struct {
	Data map[string]interface{} `json:"data"`
}

type PredictionResponse struct {
	IsAnomaly    bool    `json:"is_anomaly"`
	AnomalyScore float64 `json:"anomaly_score"`
	Prediction   string  `json:"prediction"`
	Error        string  `json:"error,omitempty"`
}

func NewClient(baseURL string, logger *zap.Logger) *Client {
	return &Client{
		baseUrl: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) Predict(ctx context.Context, requestData map[string]interface{}) (*PredictionResponse, error) {
	reqBody := PredictionRequest{Data: requestData}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("Json marshal hatası: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseUrl+"/predict", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("istek oluşturulamadı:%w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP isteği oluşturulamadı:%w", err)
	}
	defer resp.Body.Close()

	var predResp PredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&predResp); err != nil {
		return nil, fmt.Errorf("yanıt parse hatası: %w", err)
	}

	if predResp.IsAnomaly {
		c.logger.Warn("Anomali tespit edildi!",
			zap.Float64("score", predResp.AnomalyScore),
		)
	}
	return &predResp, nil
}

// func(c *Client) PredictAsync(requestData map[string]interface{},callback func(*PredictionRequest,error)) {
// 	go func () {
// 		ctx,cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 		defer cancel()
// 		result , err := c.Predict(ctx,requestData)
// 		if callback != nil {
// 			callback(result,err)
// 		}
// 	}()
// }

func (c *Client) PredictAsync(
	requestData map[string]interface{},
	callback func(*PredictionResponse, error),
) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		result, err := c.Predict(ctx, requestData)
		if callback != nil {
			callback(result, err)
		}
	}()
}

func (c *Client) HealthCheck() bool {
	resp, err := c.httpClient.Get(c.baseUrl + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// --- Dual-Model Comparison ---

type ModelResult struct {
	Model        string  `json:"model"`
	IsAnomaly    bool    `json:"is_anomaly"`
	AnomalyScore float64 `json:"anomaly_score"`
	Prediction   string  `json:"prediction"`
	InferenceMs  float64 `json:"inference_ms"`
}

type ConsensusResult struct {
	AnyAnomaly     bool     `json:"any_anomaly"`
	AllAgree       bool     `json:"all_agree"`
	FlaggedBy      []string `json:"flagged_by"`
	ModelsCompared int      `json:"models_compared"`
}

type ComparisonResponse struct {
	Models    map[string]ModelResult `json:"models"`
	Consensus ConsensusResult        `json:"consensus"`
}

type ModelStatsResponse map[string]struct {
	IsTrained        bool    `json:"is_trained"`
	TotalPredictions int     `json:"total_predictions"`
	TotalAnomalies   int     `json:"total_anomalies"`
	AvgInferenceMs   float64 `json:"avg_inference_ms"`
}

func (c *Client) PredictCompare(ctx context.Context, requestData map[string]interface{}) (*ComparisonResponse, error) {
	reqBody := PredictionRequest{Data: requestData}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("json marshal hatası: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseUrl+"/predict/compare", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("istek oluşturulamadı: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP isteği başarısız: %w", err)
	}
	defer resp.Body.Close()

	var compResp ComparisonResponse
	if err := json.NewDecoder(resp.Body).Decode(&compResp); err != nil {
		return nil, fmt.Errorf("yanıt parse hatası: %w", err)
	}

	return &compResp, nil
}

func (c *Client) PredictCompareAsync(
	requestData map[string]interface{},
	callback func(*ComparisonResponse, error),
) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := c.PredictCompare(ctx, requestData)
		if callback != nil {
			callback(result, err)
		}
	}()
}

func (c *Client) GetModelStats() (*ModelStatsResponse, error) {
	resp, err := c.httpClient.Get(c.baseUrl + "/models/stats")
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ml /models/stats: %d %s", resp.StatusCode, string(body))
	}
	var stats ModelStatsResponse
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

func (c *Client) TrainAll(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+"/train", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return fmt.Errorf("train request oluşturulamadı: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("train request başarısız: %w", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		// ML servisi eğitimi arka planda yapıyor; istatistikler hazır olana kadar bekle
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				stats, err := c.GetModelStats()
				if err != nil {
					continue
				}
				allTrained := true
				for _, s := range *stats {
					if !s.IsTrained {
						allTrained = false
						break
					}
				}
				if allTrained && len(*stats) > 0 {
					return nil
				}
			}
		}
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("train endpoint başarısız (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) EnsureModelsReady(maxRetries int) {
	if maxRetries < 1 {
		maxRetries = 1
	}
	for i := 0; i < maxRetries; i++ {
		stats, err := c.GetModelStats()
		if err == nil {
			if len(*stats) > 0 {
				allTrained := true
				for _, s := range *stats {
					if !s.IsTrained {
						allTrained = false
						break
					}
				}
				if allTrained {
					c.logger.Info("ML modelleri hazır", zap.Int("models", len(*stats)))
					return
				}
			}
			c.logger.Warn("ML modelleri eğitimsiz görünüyor, /train çağrılıyor", zap.Int("attempt", i+1))
		} else {
			c.logger.Warn("ML model stats okunamadı, /train denenecek", zap.Int("attempt", i+1), zap.Error(err))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
		err = c.TrainAll(ctx)
		cancel()
		if err != nil {
			c.logger.Warn("ML train çağrısı başarısız", zap.Int("attempt", i+1), zap.Error(err))
			time.Sleep(3 * time.Second)
			continue
		}
		time.Sleep(2 * time.Second)
	}
	c.logger.Warn("ML modelleri otomatik hazır hale getirilemedi; servis eğitimsiz olabilir")
}
