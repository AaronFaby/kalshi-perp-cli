package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()
	fn()
	_ = w.Close()
	os.Stdout = old
	return <-done
}

func TestOrdersCreateDryRun_EmitsRequiredJSONFields(t *testing.T) {
	root := NewRoot()
	root.SetArgs([]string{
		"orders", "create",
		"--ticker", "KXBTCPERP1",
		"--side", "bid",
		"--count", "1.00",
		"--price", "6.2000",
		"--tif", "good_till_canceled",
		"--stp", "maker",
		"--dry-run",
		"--format", "json",
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
	for _, k := range []string{"ticker", "client_order_id", "side", "count", "price", "time_in_force", "self_trade_prevention_type"} {
		if _, ok := body[k]; !ok {
			t.Errorf("missing %s in %s", k, stdout)
		}
	}
	if body["ticker"] != "KXBTCPERP1" || body["side"] != "bid" {
		t.Fatalf("%v", body)
	}
	if _, ok := body["count"].(string); !ok {
		t.Fatalf("count %T", body["count"])
	}
	if _, ok := body["price"].(string); !ok {
		t.Fatalf("price %T", body["price"])
	}
	if body["client_order_id"] == "" {
		t.Fatal("expected auto client_order_id")
	}
}

func TestOrdersCreate_RequiresFlags(t *testing.T) {
	root := NewRoot()
	root.SetArgs([]string{"orders", "create", "--ticker", "X", "--format", "json"})
	var errBuf bytes.Buffer
	root.SetErr(&errBuf)
	root.SetOut(&errBuf)
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Fatal(err)
	}
}

func TestRootHasCoreCommands(t *testing.T) {
	root := NewRoot()
	want := []string{
		"account", "auth", "balance", "exchange", "fees", "fills", "funding",
		"groups", "markets", "orders", "positions", "risk", "stream", "subaccounts", "transfer", "version",
	}
	for _, name := range want {
		found := false
		for _, c := range root.Commands() {
			if c.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing command %s", name)
		}
	}
}
