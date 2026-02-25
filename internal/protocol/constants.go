package protocol

import "time"

const (
	ControlPort = ":7000"

	ProxyPort = ":8085"

	HandshakeTimeout = 10 * time.Second
	ReadTimeout      = 30 * time.Second
	WriteTimeout     = 30 * time.Second
)

const (
	MsgTypeRegister = "REGISTER"

	MsgTypeRegistered = "REGISTERED"
	MsgTypeError      = "ERROR"

	MsgTypeRegisterTCP = "REGISTER_TCP"
	MsgTypeRegisterUDP = "REGISTER_UDP"

	MsgTypePing = "PING"
	MsgTypePong = "PONG"
)

const (
	BaseDomain = "gorenel.io"

	SubdomainLength = 12
)
