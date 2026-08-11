package api

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New(srv.URL+"/trade-api/v2", "test-key-id", key, 5*time.Second)
}

func TestDo_SignsAndGETs(t *testing.T) {
	var gotKey, gotSig, gotTs, gotPath string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("KALSHI-ACCESS-KEY")
		gotSig = r.Header.Get("KALSHI-ACCESS-SIGNATURE")
		gotTs = r.Header.Get("KALSHI-ACCESS-TIMESTAMP")
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]any{"enabled": true})
	})

	var out GetMarginEnabledResponse
	if err := c.Get(context.Background(), "/margin/enabled", nil, &out); err != nil {
		t.Fatal(err)
	}
	if !out.Enabled {
		t.Fatal("expected enabled")
	}
	if gotKey != "test-key-id" || gotSig == "" || gotTs == "" {
		t.Fatalf("headers key=%q sig=%q ts=%q", gotKey, gotSig, gotTs)
	}
	if gotPath != "/trade-api/v2/margin/enabled" {
		t.Fatalf("path %s", gotPath)
	}
}

func TestDo_QueryNotInSignPath(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("ticker") != "ABC" {
			t.Errorf("query ticker=%s", r.URL.Query().Get("ticker"))
		}
		_ = json.NewEncoder(w).Encode(GetMarginOrdersResponse{Orders: []MarginOrder{}, Cursor: ""})
	})
	_, err := c.GetMarginOrders(context.Background(), OrdersParams{Ticker: "ABC"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDo_APIError(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Code: "authentication_error", Message: "bad key"})
	})
	err := c.Get(context.Background(), "/margin/balance", nil, &GetMarginBalanceResponse{})
	if err == nil {
		t.Fatal("expected error")
	}
	ae, ok := err.(*APIError)
	if !ok {
		t.Fatalf("type %T", err)
	}
	if ae.StatusCode != 401 || !strings.Contains(ae.Error(), "bad key") {
		t.Fatal(ae)
	}
}

func TestCreateOrder_PostsBody(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method %s", r.Method)
		}
		var req CreateMarginOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Ticker != "X" || req.Side != "bid" || req.Count != "1.00" {
			t.Fatalf("%+v", req)
		}
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(CreateMarginOrderResponse{
			OrderID: "oid", FillCount: "0.00", RemainingCount: "1.00", ClientOrderID: req.ClientOrderID,
		})
	})
	resp, err := c.CreateMarginOrder(context.Background(), CreateMarginOrderRequest{
		Ticker: "X", ClientOrderID: "c1", Side: "bid", Count: "1.00", Price: "0.50",
		TimeInForce: "good_till_canceled", SelfTradePreventionType: "taker_at_cross",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.OrderID != "oid" {
		t.Fatal(resp)
	}
}
