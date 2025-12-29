package factory

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func RandomString() string {
	return uuid.NewString()
}

func randomEmail() string {
	return uuid.NewString() + "@example.com"
}

func marshal(t *testing.T, data any) []byte {
	bytes, err := json.Marshal(data)
	require.NoError(t, err, "expected no error marshaling data")
	return bytes
}

func RandomDigits(length int) string {
	digits := make([]byte, length)
	for i := range digits {
		digits[i] = byte('0' + rand.Intn(10))
	}
	return string(digits)
}

func randomSha256() string {
	randStr := uuid.NewString()
	sha256_ := sha256.Sum256([]byte(randStr))
	return hex.EncodeToString(sha256_[:])
}
