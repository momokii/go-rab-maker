package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// The GenerateTokenSession function generates a random token of 32 bytes (256 bits) and returns it as
// a hexadecimal string.
func GenerateTokenSession() (string, error) {
	b := make([]byte, 32) // 32 bytes = 256 bits
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func GenerateUUIDV4() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Set the version to 4 (randomly generated UUID)
	b[6] = (b[6] & 0x0f) | 0x40
	// Set the variant to RFC 4122
	b[8] = (b[8] & 0x3f) | 0x80

	uuid := hex.EncodeToString(b[0:4]) + "-" +
		hex.EncodeToString(b[4:6]) + "-" +
		hex.EncodeToString(b[6:8]) + "-" +
		hex.EncodeToString(b[8:10]) + "-" +
		hex.EncodeToString(b[10:16])

	return uuid, nil
}
