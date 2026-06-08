package server

import (
	"fmt"
)

type OpenAIProvider struct {
	Name   string
	APIKey string
	Model  string
}

func (p *OpenAIProvider) GetName() string {
	return p.Name
}

func (p *OpenAIProvider) GetURL() string {
	return "https://api.openai.com/v1/chat/completions"
}

func (p *OpenAIProvider) GetAuthHeader() string {
	return fmt.Sprintf("Bearer %s", p.APIKey)
}

type AnthropicProvider struct {
	Name   string
	APIKey string
	Model  string
}

func (p *AnthropicProvider) GetName() string {
	return p.Name
}

func (p *AnthropicProvider) GetURL() string {
	return "https://api.anthropic.com/v1/messages"
}

func (p *AnthropicProvider) GetAuthHeader() string {
	return p.APIKey // Anthropic use x-api-key header usually handled in proxy setup if specialized
}
