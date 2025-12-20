package protocol

import "time"

const (
	ControlPort = ":7000"

	ProxyPort = ":8080"

	HandshakeTimeout = 10 * time.Second
	ReadTimeout      = 30 * time.Second
	WriteTimeout     = 30 * time.Second
)

const (
	MsgTypeRegister = "REGISTER"

	MsgTypeRegistered = "REGISTERED"
	MsgTypeError      = "ERROR"

	MsgTypePing = "PING"
	MsgTypePong = "PONG"
)

const (
	BaseDomain = "tunnel.local"

	SubdomainLength = 12
)
