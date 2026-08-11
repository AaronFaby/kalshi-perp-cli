package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"
)

// LoadPrivateKey reads a PEM-encoded RSA private key from path.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	return ParsePrivateKey(data)
}

// ParsePrivateKey parses PEM bytes (PKCS#1 or PKCS#8) into an RSA private key.
func ParsePrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("private key: no PEM block found")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("private key: parse PKCS#1/PKCS#8: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key: not RSA")
	}
	return key, nil
}

// SignRequest produces the KALSHI-ACCESS-SIGNATURE value.
// message = timestamp + METHOD + pathWithoutQuery
// path must be the full URL path (e.g. /trade-api/v2/margin/orders), without query string.
func SignRequest(privateKey *rsa.PrivateKey, timestamp, method, path string) (string, error) {
	path = strings.SplitN(path, "?", 2)[0]
	message := timestamp + strings.ToUpper(method) + path
	hash := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, hash[:], &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
		Hash:       crypto.SHA256,
	})
	if err != nil {
		return "", fmt.Errorf("sign request: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// TimestampMS returns the current time as a millisecond epoch string.
func TimestampMS() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}

// Headers builds the three Kalshi auth headers for a request.
func Headers(keyID string, privateKey *rsa.PrivateKey, method, path string) (map[string]string, error) {
	ts := TimestampMS()
	sig, err := SignRequest(privateKey, ts, method, path)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"KALSHI-ACCESS-KEY":       keyID,
		"KALSHI-ACCESS-TIMESTAMP": ts,
		"KALSHI-ACCESS-SIGNATURE": sig,
	}, nil
}
