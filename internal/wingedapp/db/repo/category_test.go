package repo

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	var hashed [32]byte

	type shaed struct {
		Email  string `json:"email"`
		Number string `json:"number"`
	}

	a := shaed{
		Email:  "abe",
		Number: "hello",
	}

	b, err := json.Marshal(a)
	require.NoError(t, err)

	fmt.Println("==== b:", string(b))

	hashed = sha256.Sum256(b)
	var converted []byte
	converted = hashed[:]
	fmt.Println("converted:", converted)
}
