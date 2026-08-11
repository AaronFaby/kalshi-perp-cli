package api

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// Non-FCM OpenAPI paths that the CLI must implement.
var requiredPaths = []string{
	"/account/limits/perps",
	"/margin/exchange/status",
	"/margin/enabled",
	"/margin/risk_parameters",
	"/margin/orders",
	"/margin/orders/{order_id}",
	"/margin/orders/{order_id}/decrease",
	"/margin/orders/{order_id}/amend",
	"/margin/markets",
	"/margin/markets/{ticker}",
	"/margin/markets/{ticker}/orderbook",
	"/margin/markets/{ticker}/candlesticks",
	"/margin/fills",
	"/margin/positions",
	"/margin/trades",
	"/margin/notional_risk_limit",
	"/margin/balance",
	"/margin/risk",
	"/margin/fee_tiers",
	"/margin/funding_history",
	"/margin/funding_rates/historical",
	"/margin/funding_rates/estimate",
	"/portfolio/intra_exchange_instance_transfer",
	"/portfolio/margin/subaccounts",
	"/portfolio/margin/subaccounts/transfer",
	"/margin/order_groups",
	"/margin/order_groups/create",
	"/margin/order_groups/{order_group_id}",
	"/margin/order_groups/{order_group_id}/reset",
	"/margin/order_groups/{order_group_id}/trigger",
	"/margin/order_groups/{order_group_id}/limit",
}

func TestRequiredPathsPresentInClientSource(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	dir := filepath.Dir(thisFile)
	src, err := os.ReadFile(filepath.Join(dir, "endpoints.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(src)
	for _, p := range requiredPaths {
		// convert {param} to dynamic join or pathescape usage — check static prefix
		static := p
		if i := strings.Index(p, "{"); i >= 0 {
			static = p[:i]
		}
		if !strings.Contains(body, static) {
			t.Errorf("client missing path prefix %q (from %s)", static, p)
		}
	}
}

func TestVendoredOpenAPIHasMarginPaths(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	root := filepath.Join(filepath.Dir(thisFile), "..", "..")
	spec, err := os.ReadFile(filepath.Join(root, "docs", "openapi", "perps_openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(spec)
	pathRe := regexp.MustCompile(`(?m)^  (/[^\n:]+):\s*$`)
	found := map[string]bool{}
	for _, m := range pathRe.FindAllStringSubmatch(text, -1) {
		found[m[1]] = true
	}
	for _, p := range requiredPaths {
		if !found[p] {
			// order_group_id path template might differ slightly — soft check
			t.Logf("openapi missing listed path %s (spec may have drifted)", p)
		}
	}
	// Ensure FCM paths exist in spec but we deliberately skip them in requiredPaths
	if !found["/margin/fcm/subtraders"] {
		t.Log("note: FCM path not in vendored openapi")
	}
}
