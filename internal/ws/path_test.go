package ws

import "testing"

func TestWsSignPath(t *testing.T) {
	cases := map[string]string{
		"wss://external-api-margin-ws.demo.kalshi.co/trade-api/ws/v2/margin": "/trade-api/ws/v2/margin",
		"wss://host/trade-api/ws/v2/margin":                                 "/trade-api/ws/v2/margin",
	}
	for in, want := range cases {
		if got := wsSignPath(in); got != want {
			t.Fatalf("%s: got %s want %s", in, got, want)
		}
	}
}
