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
	ClientID        string `json:"client_id"`
	Version         string `json:"version"`
	APIKey          string `json:"api_key"`
	CustomSubdomain string `json:"custom_subdomain,omitempty"`
	CustomDomain    string `json:"custom_domain,omitempty"`
	TunnelType      string `json:"tunnel_type,omitempty"`
	LocalPort       int    `json:"local_port,omitempty"`
	// Security policies (applies to HTTP/WSS tunnel traffic)
	KeyAuthToken string   `json:"key_auth_token,omitempty"`
	IPWhitelist  []string `json:"ip_whitelist,omitempty"`
	CORSEnabled  bool     `json:"cors_enabled,omitempty"`
	Auth         string   `json:"auth,omitempty"`
	// PublicAccess disables default KeyAuth isolation.
	// When explicitly false, tunnels are protected with an auto-generated token.
	PublicAccess *bool `json:"public_access,omitempty"`
}

type RegisterResponse struct {
	Subdomain   string `json:"subdomain"`
	FullURL     string `json:"full_url"`
	PublicPort  int    `json:"public_port,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
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
		LocalPort:    80, // Default or should be passed
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
