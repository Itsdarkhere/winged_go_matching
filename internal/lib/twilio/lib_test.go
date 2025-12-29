package twilio_test

import (
	"context"
	"os"
	"testing"
	"wingedapp/pgtester/internal/lib/twilio"

	"github.com/stretchr/testify/require"
)

const (
	envPrefix    = "TWILIO_TEST"
	accountSID   = "ACCOUNT_SID"
	authToken    = "AUTH_TOKEN"
	sendToNumber = "SEND_TO_NUMBER"
	sentFrom     = "SENT_FROM"
)

func env(key string) string {
	return os.Getenv(envPrefix + "_" + key)
}

func hasCreds() (ok bool, missing []string) {
	if env(accountSID) == "" {
		missing = append(missing, envPrefix+"_"+accountSID)
	}
	if env(authToken) == "" {
		missing = append(missing, envPrefix+"_"+authToken)
	}
	if env(sendToNumber) == "" {
		missing = append(missing, envPrefix+"_"+sendToNumber)
	}
	return len(missing) == 0, missing
}

func autoskip(t testing.TB) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	if ok, missing := hasCreds(); !ok {
		t.Skipf("skipping: missing env vars: %v", missing)
	}
}

func TestClient_SendMessage(t *testing.T) {
	t.Skip() // no need to test this.. all is good
	autoskip(t)

	client, err := twilio.New(&twilio.Config{
		AccountSID: env(accountSID),
		AuthToken:  env(authToken),
		From:       env(sentFrom),
	})
	require.NoError(t, err, "expected no error creating Twilio client")

	err = client.SendMessage(context.TODO(), env(sendToNumber), "Hello!")
	require.NoError(t, err, "expected no error sending message")
}
