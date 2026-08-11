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

func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

func TestRequiredPathsPresentInClientSource(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(repoRoot(t), "internal", "api", "endpoints.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(src)
	for _, p := range requiredPaths {
		static := p
		if i := strings.Index(p, "{"); i >= 0 {
			static = p[:i]
		}
		if !strings.Contains(body, static) {
			t.Errorf("client missing path prefix %q (from %s)", static, p)
		}
	}
}

func TestCLICommandTreeCoversNonFCMOperations(t *testing.T) {
	// Map OpenAPI operation intent -> required cobra Use fragment(s) under internal/cli.
	// These strings must appear in CLI sources so subcommands stay wired.
	// Substrings that must appear in CLI sources (Use strings / flags).
	needles := []string{
		"account",
		"limits",
		"exchange",
		"enabled",
		"markets",
		"orderbook",
		"candles",
		"trades",
		"orders",
		"cancel",
		"amend",
		"decrease",
		"positions",
		"fills",
		"balance",
		"risk",
		"notional-limit",
		"fees",
		"tiers",
		"funding",
		"estimate",
		"subaccounts",
		"groups",
		"stream",
		"compute-available-balance",
		"amount-cents",
		"client-transfer-id",
		"dry-run",
	}
	cliDir := filepath.Join(repoRoot(t), "internal", "cli")
	entries, err := os.ReadDir(cliDir)
	if err != nil {
		t.Fatal(err)
	}
	var body strings.Builder
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(cliDir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		body.Write(b)
	}
	src := body.String()
	for _, n := range needles {
		if !strings.Contains(src, n) {
			t.Errorf("CLI sources missing %s", n)
		}
	}
}

func TestVendoredOpenAPIHasAllRequiredNonFCMPaths(t *testing.T) {
	spec, err := os.ReadFile(filepath.Join(repoRoot(t), "docs", "openapi", "perps_openapi.yaml"))
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
			t.Errorf("openapi missing listed path %s", p)
		}
	}
	// FCM must remain in spec but not in requiredPaths
	if !found["/margin/fcm/subtraders"] {
		t.Log("note: FCM path not in vendored openapi")
	}
	// Ensure we intentionally exclude FCM from requiredPaths
	for _, p := range requiredPaths {
		if strings.Contains(p, "/fcm/") {
			t.Errorf("requiredPaths must not include FCM: %s", p)
		}
	}
}

func TestNonFCMOpenAPIOpsMatchClientPathCount(t *testing.T) {
	spec, err := os.ReadFile(filepath.Join(repoRoot(t), "docs", "openapi", "perps_openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	// Count non-FCM path keys
	pathRe := regexp.MustCompile(`(?m)^  (/[^\n:]+):\s*$`)
	comp := regexp.MustCompile(`(?m)^components:`)
	loc := comp.FindIndex(spec)
	section := string(spec)
	if loc != nil {
		section = string(spec[:loc[0]])
	}
	var nonFCM int
	for _, m := range pathRe.FindAllStringSubmatch(section, -1) {
		if !strings.Contains(m[1], "/fcm/") {
			nonFCM++
		}
	}
	if nonFCM < len(requiredPaths) {
		t.Fatalf("openapi non-FCM paths %d < requiredPaths %d", nonFCM, len(requiredPaths))
	}
}
