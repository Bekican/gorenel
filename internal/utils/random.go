package utils

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

func GenerateSubDomain(length int) string {
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			result[i] = charset[0]
			continue
		}
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

func GenerateClientID() string {
	return GenerateSubDomain(16)
}

func GenerateTunnelToken() string {
	// Token is used as X-TOKEN header value.
	// Keep it URL-safe and easy to paste.
	return "tt_" + GenerateSubDomain(32)
}
