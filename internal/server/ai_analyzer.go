package server

import (
	"encoding/json"
	"strings"
)

// AIAnalyzer extracts LLM specific metadata from request/response bodies
type AIAnalyzer struct{}

func NewAIAnalyzer() *AIAnalyzer {
	return &AIAnalyzer{}
}

// AnalyzeRequest identifies the AI provider and extracts prompt info if possible
func (a *AIAnalyzer) AnalyzeRequest(host string, path string, body []byte) *AIMetadata {
	// Detect OpenAI
	if strings.Contains(host, "openai.com") || strings.Contains(path, "/v1/chat/completions") {
		return a.parseOpenAIRequest(body)
	}

	// Detect Anthropic
	if strings.Contains(host, "anthropic.com") || strings.Contains(path, "/v1/messages") {
		return a.parseAnthropicRequest(body)
	}

	return nil
}

func (a *AIAnalyzer) parseOpenAIRequest(body []byte) *AIMetadata {
	var req struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		return nil
	}

	meta := &AIMetadata{
		Provider: "OpenAI",
		Model:    req.Model,
	}

	// Extract prompt from last message or all messages
	var promptBuilder strings.Builder
	for _, m := range req.Messages {
		promptBuilder.WriteString(m.Role + ": " + m.Content + "\n")
	}
	meta.Prompt = promptBuilder.String()

	// Phase 5: Detection
	a.DetectInjection(meta)

	return meta
}

func (a *AIAnalyzer) parseAnthropicRequest(body []byte) *AIMetadata {
	var req struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		return nil
	}

	meta := &AIMetadata{
		Provider: "Anthropic",
		Model:    req.Model,
	}

	var promptBuilder strings.Builder
	for _, m := range req.Messages {
		promptBuilder.WriteString(m.Role + ": " + m.Content + "\n")
	}
	meta.Prompt = promptBuilder.String()

	// Phase 5: Detection
	a.DetectInjection(meta)

	return meta
}

// AnalyzeResponse parses completion and token usage from response bodies
func (a *AIAnalyzer) AnalyzeResponse(meta *AIMetadata, body []byte) {
	if meta == nil {
		return
	}

	switch meta.Provider {
	case "OpenAI":
		a.parseOpenAIResponse(meta, body)
	case "Anthropic":
		a.parseAnthropicResponse(meta, body)
	}
}

func (a *AIAnalyzer) parseOpenAIResponse(meta *AIMetadata, body []byte) {
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return
	}

	if len(resp.Choices) > 0 {
		meta.Completion = resp.Choices[0].Message.Content
	}
	meta.Tokens.Prompt = resp.Usage.PromptTokens
	meta.Tokens.Completion = resp.Usage.CompletionTokens
	meta.Tokens.Total = resp.Usage.TotalTokens
}

func (a *AIAnalyzer) parseAnthropicResponse(meta *AIMetadata, body []byte) {
	var resp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return
	}

	if len(resp.Content) > 0 {
		meta.Completion = resp.Content[0].Text
	}
	meta.Tokens.Prompt = resp.Usage.InputTokens
	meta.Tokens.Completion = resp.Usage.OutputTokens
	meta.Tokens.Total = resp.Usage.InputTokens + resp.Usage.OutputTokens
}

// DetectInjection performs heuristic analysis for prompt injection attempts
func (a *AIAnalyzer) DetectInjection(meta *AIMetadata) {
	if meta == nil || meta.Prompt == "" {
		return
	}

	prompt := strings.ToLower(meta.Prompt)

	type weightedPattern struct {
		pattern string
		weight  float64
	}

	patterns := []weightedPattern{
		// Critical (0.50) — direct manipulation attempts
		{"ignore previous instructions", 0.50},
		{"ignore all previous", 0.50},
		{"jailbreak", 0.50},
		{"bypass safety", 0.50},
		{"bypass filters", 0.50},
		{"önceki talimatları unut", 0.50},
		{"tüm talimatları yoksay", 0.50},
		{"güvenlik filtrelerini atla", 0.50},

		// High (0.35) — prompt extraction & role override
		{"system prompt", 0.35},
		{"reveal your rules", 0.35},
		{"developer mode", 0.35},
		{"dan mode", 0.35},
		{"what are your instructions", 0.35},
		{"repeat everything above", 0.35},
		{"show me your prompt", 0.35},
		{"output your rules", 0.35},
		{"sistem promptunu göster", 0.35},
		{"kurallarını göster", 0.35},

		// Medium (0.20) — social engineering & roleplay
		{"you are now", 0.20},
		{"pretend you are", 0.20},
		{"act as if", 0.20},
		{"roleplay as", 0.20},
		{"stay out of character", 0.20},
		{"do not mention", 0.20},
		{"base64 decode", 0.20},
		{"translate from hex", 0.20},
		{"şimdi sen bir", 0.20},
	}

	var totalScore float64
	var matched []string

	for _, p := range patterns {
		if strings.Contains(prompt, p.pattern) {
			totalScore += p.weight
			matched = append(matched, p.pattern)
		}
	}

	if len(matched) > 0 {
		meta.IsSecurityRisk = true
		if totalScore > 1.0 {
			totalScore = 1.0
		}
		meta.RiskScore = totalScore
		meta.RiskReason = "Prompt Injection: " + strings.Join(matched, ", ")
	}
}
