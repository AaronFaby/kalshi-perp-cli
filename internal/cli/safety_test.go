package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aaronfaby/kalshi-perp-cli/internal/config"
)

func isolatedRoot(t *testing.T, args ...string) (*bytes.Buffer, error) {
	t.Helper()
	root := NewRoot()
	cfg := filepath.Join(t.TempDir(), "missing.yaml")
	root.SetArgs(append([]string{"--config", cfg}, args...))
	var errBuf bytes.Buffer
	root.SetErr(&errBuf)
	root.SetOut(&errBuf)
	return &errBuf, root.Execute()
}

func TestProdMutationRequiresConfirm(t *testing.T) {
	_, err := isolatedRoot(t, "--env", "prod", "orders", "cancel", "oid")
	if err == nil || !strings.Contains(err.Error(), "confirm-prod") {
		t.Fatalf("got %v", err)
	}
}

func TestProdHostOverrideRequiresConfirm(t *testing.T) {
	_, err := isolatedRoot(t, "--env", "demo", "--base-url", config.DefaultBaseURL(config.EnvProd), "orders", "cancel", "oid")
	if err == nil || !strings.Contains(err.Error(), "confirm-prod") {
		t.Fatalf("got %v", err)
	}
}

func TestDemoMutationDoesNotRequireConfirm(t *testing.T) {
	_, err := isolatedRoot(t, "--env", "demo", "orders", "cancel", "oid")
	if err == nil {
		t.Fatal("expected auth error")
	}
	if strings.Contains(err.Error(), "confirm-prod") {
		t.Fatal(err)
	}
	if !strings.Contains(err.Error(), "api key") {
		t.Fatal(err)
	}
}

func TestProdConfirmPrintsWarningThenRequiresAuth(t *testing.T) {
	errBuf, err := isolatedRoot(t, "--env", "prod", "--confirm-prod", "orders", "cancel", "oid")
	if err == nil {
		t.Fatal("expected auth error")
	}
	if strings.Contains(err.Error(), "confirm-prod") {
		t.Fatal(err)
	}
	if !strings.Contains(errBuf.String(), "warning: targeting production") {
		t.Fatalf("stderr=%q err=%v", errBuf.String(), err)
	}
}

func TestProdDryRunDoesNotRequireConfirm(t *testing.T) {
	root := NewRoot()
	cfg := filepath.Join(t.TempDir(), "missing.yaml")
	root.SetArgs([]string{
		"--config", cfg, "--env", "prod", "--format", "json",
		"orders", "create",
		"--ticker", "KXBTCPERP1",
		"--side", "bid",
		"--count", "1.00",
		"--price", "6.2000",
		"--tif", "good_till_canceled",
		"--stp", "maker",
		"--dry-run",
	})
	var errBuf bytes.Buffer
	root.SetErr(&errBuf)
	stdout := captureStdout(t, func() {
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	var body map[string]any
	if err := json.Unmarshal([]byte(stdout), &body); err != nil {
		t.Fatalf("json: %v\nstdout=%q stderr=%q", err, stdout, errBuf.String())
	}
	if body["ticker"] != "KXBTCPERP1" {
		t.Fatalf("%v", body)
	}
}

func TestTransferExchangeRejectsBothAmountFlags(t *testing.T) {
	_, err := isolatedRoot(t,
		"transfer", "exchange",
		"--source", "event_contract",
		"--destination", "margined",
		"--amount-centicents", "50000",
		"--amount-dollars", "1.00",
	)
	if err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("got %v", err)
	}
}

func TestTransferExchangeRequiresAnAmountFlag(t *testing.T) {
	_, err := isolatedRoot(t,
		"transfer", "exchange",
		"--source", "event_contract",
		"--destination", "margined",
	)
	if err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("got %v", err)
	}
}
