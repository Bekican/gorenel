package protocol

import (
	"encoding/json"
	"io"
)

type Message struct {
	Type    string `json:"type"`
	Payload string `json:"payload,omitempty"`
}

type RegisterRequest struct {
	ClientID     string `json:"client_id"`
	Version      string `json:"version"`
	APIKey       string `json:"api_key"`
	CustomDomain string `json:"custom_domain,omitempty"` // --- NEW: Custom Domain field ---
}

type RegisterResponse struct {
	Subdomain string `json:"subdomain"`
	FullURL   string `json:"full_url"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func WriteMessage(w io.Writer, msg Message) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(msg)
}

func ReadMessage(r io.Reader) (*Message, error) {
	var msg Message
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// NewRegisterMessage now takes an optional custom domain
func NewRegisterMessage(clientId, version, apiKey, domain string) Message {
	req := RegisterRequest{
		ClientID:     clientId,
		Version:      version,
		APIKey:       apiKey,
		CustomDomain: domain,
	}
	payload, _ := json.Marshal(req)

	return Message{
		Type:    MsgTypeRegister,
		Payload: string(payload),
	}
}

func NewRegisteredMessage(subdomain, fullURL string) Message {
	resp := RegisterResponse{
		Subdomain: subdomain,
		FullURL:   fullURL,
	}
	payload, _ := json.Marshal(resp)

	return Message{
		Type:    MsgTypeRegistered,
		Payload: string(payload),
	}
}

func NewErrorMessage(code int, message string) Message {
	errResp := ErrorResponse{
		Code:    code,
		Message: message,
	}
	payload, _ := json.Marshal(errResp)

	return Message{
		Type:    MsgTypeError,
		Payload: string(payload),
	}
}
