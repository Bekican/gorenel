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
