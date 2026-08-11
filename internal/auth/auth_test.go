package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"
)

func testKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func TestSignRequest_StripsQueryAndUppercasesMethod(t *testing.T) {
	key := testKey(t)
	ts := "1703123456789"
	path := "/trade-api/v2/margin/orders?limit=5"
	sig, err := SignRequest(key, ts, "get", path)
	if err != nil {
		t.Fatal(err)
	}
	if sig == "" {
		t.Fatal("empty signature")
	}

	// Verify with the same message construction.
	message := ts + "GET" + "/trade-api/v2/margin/orders"
	hash := sha256.Sum256([]byte(message))
	raw, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		t.Fatal(err)
	}
	if err := rsa.VerifyPSS(&key.PublicKey, crypto.SHA256, hash[:], raw, &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
		Hash:       crypto.SHA256,
	}); err != nil {
		t.Fatalf("verify failed: %v", err)
	}
}

func TestParsePrivateKey_PKCS1(t *testing.T) {
	key := testKey(t)
	der := x509.MarshalPKCS1PrivateKey(key)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
	parsed, err := ParsePrivateKey(pemBytes)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.N.Cmp(key.N) != 0 {
		t.Fatal("key mismatch")
	}
}

func TestParsePrivateKey_PKCS8(t *testing.T) {
	key := testKey(t)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	parsed, err := ParsePrivateKey(pemBytes)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.N.Cmp(key.N) != 0 {
		t.Fatal("key mismatch")
	}
}

func TestHeaders(t *testing.T) {
	key := testKey(t)
	h, err := Headers("key-id", key, "GET", "/trade-api/v2/margin/balance")
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"KALSHI-ACCESS-KEY", "KALSHI-ACCESS-TIMESTAMP", "KALSHI-ACCESS-SIGNATURE"} {
		if h[k] == "" {
			t.Fatalf("missing header %s", k)
		}
	}
	if h["KALSHI-ACCESS-KEY"] != "key-id" {
		t.Fatal(h["KALSHI-ACCESS-KEY"])
	}
	if !strings.HasPrefix(h["KALSHI-ACCESS-TIMESTAMP"], "1") && len(h["KALSHI-ACCESS-TIMESTAMP"]) < 10 {
		t.Fatalf("bad timestamp: %s", h["KALSHI-ACCESS-TIMESTAMP"])
	}
}
