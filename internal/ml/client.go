package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
