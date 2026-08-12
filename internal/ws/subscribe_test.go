package ws

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBuildSubscribeMessage_ChannelsAndTickers(t *testing.T) {
	b, err := BuildSubscribeMessage(1, SubscribeParams{
		Channels: []string{"ticker", "trade"},
		Tickers:  []string{"KXBTCPERP1", "KXETHPERP1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var msg map[string]any
	if err := json.Unmarshal(b, &msg); err != nil {
		t.Fatal(err)
	}
	if msg["cmd"] != "subscribe" {
		t.Fatalf("%v", msg)
	}
	params := msg["params"].(map[string]any)
	chs := params["channels"].([]any)
	if len(chs) != 2 || chs[0] != "ticker" {
		t.Fatalf("%v", chs)
	}
	ticks := params["market_tickers"].([]any)
	if len(ticks) != 2 {
		t.Fatalf("%v", ticks)
	}
}

func TestBuildSubscribeMessage_SingleTicker(t *testing.T) {
	b, err := BuildSubscribeMessage(2, SubscribeParams{
		Channels: []string{"orderbook_delta"},
		Tickers:  []string{"KXBTCPERP1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var msg map[string]any
	if err := json.Unmarshal(b, &msg); err != nil {
		t.Fatal(err)
	}
	params := msg["params"].(map[string]any)
	if params["market_ticker"] != "KXBTCPERP1" {
		t.Fatalf("%v", params)
	}
	if _, ok := params["market_tickers"]; ok {
		t.Fatal("should not set market_tickers for single ticker")
	}
}

func TestBuildSubscribeMessage_DefaultChannel(t *testing.T) {
	b, err := BuildSubscribeMessage(3, SubscribeParams{})
	if err != nil {
		t.Fatal(err)
	}
	var msg map[string]any
	_ = json.Unmarshal(b, &msg)
	params := msg["params"].(map[string]any)
	chs := params["channels"].([]any)
	if len(chs) != 1 || chs[0] != "ticker" {
		t.Fatalf("%v", chs)
	}
}

func TestWsSignPath_NoIdleTimeoutInSource(t *testing.T) {
	if got := wsSignPath("wss://external-api-margin-ws.demo.kalshi.co/trade-api/ws/v2/margin"); got != "/trade-api/ws/v2/margin" {
		t.Fatal(got)
	}
}

func TestRunSourceUsesParentContextForReads(t *testing.T) {
	// Regression: idle streams must not use a per-read timeout (closes quiet connections).
	src, err := os.ReadFile("client.go")
	if err != nil {
		t.Fatal(err)
	}
	body := string(src)
	if strings.Contains(body, "conn.Read(context.WithTimeout") {
		t.Fatal("ws client must not impose per-read WithTimeout")
	}
	if !strings.Contains(body, "conn.Read(sessionCtx)") {
		t.Fatal("expected session-context reads")
	}
	if !strings.Contains(body, "conn.Ping(") {
		t.Fatal("expected idle ping watchdog")
	}
}

func TestNextWSBackoffResetsAndCaps(t *testing.T) {
	if got := nextWSBackoff(time.Second); got != 2*time.Second {
		t.Fatal(got)
	}
	if got := nextWSBackoff(20 * time.Second); got != maxWSBackoff {
		t.Fatal(got)
	}
	if got := nextWSBackoff(maxWSBackoff); got != maxWSBackoff {
		t.Fatal(got)
	}
}
