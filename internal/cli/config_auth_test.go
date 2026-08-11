package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigShowRedactsShortAPIKey(t *testing.T) {
	t.Setenv("KALSHI_API_KEY", "abc")
	cmd := newConfigCmd(&RootOptions{ConfigPath: filepath.Join(t.TempDir(), "missing.yaml")})
	cmd.SetArgs([]string{"show"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "abc") {
		t.Fatalf("API key leaked: %s", out.String())
	}
}
