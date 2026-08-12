package api

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
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

func TestDo_SignsTimestampMethodAndFullPath(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts := r.Header.Get("KALSHI-ACCESS-TIMESTAMP")
		sig := r.Header.Get("KALSHI-ACCESS-SIGNATURE")
		raw, err := base64.StdEncoding.DecodeString(sig)
		if err != nil {
			t.Fatal(err)
		}
		message := ts + "GET" + "/trade-api/v2/margin/enabled"
		hash := sha256.Sum256([]byte(message))
		if err := rsa.VerifyPSS(&key.PublicKey, crypto.SHA256, hash[:], raw, &rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthEqualsHash,
			Hash:       crypto.SHA256,
		}); err != nil {
			t.Fatalf("signed message was not timestamp+METHOD+/trade-api/v2/...: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"enabled": true})
	}))
	t.Cleanup(srv.Close)
	c := New(srv.URL+"/trade-api/v2", "test-key-id", key, 5*time.Second)
	if err := c.Get(context.Background(), "/margin/enabled", nil, &GetMarginEnabledResponse{}); err != nil {
		t.Fatal(err)
	}
}

func TestDo_RefusesRedirects(t *testing.T) {
	var followed bool
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/stolen") {
			followed = true
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/trade-api/v2/stolen", http.StatusFound)
	})
	err := c.Get(context.Background(), "/margin/enabled", nil, &GetMarginEnabledResponse{})
	if err == nil {
		t.Fatal("expected redirect error")
	}
	if followed {
		t.Fatal("followed redirect")
	}
	if !strings.Contains(err.Error(), "redirect") {
		t.Fatal(err)
	}
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

func TestTransferBetweenSubaccounts_PostsSchemaFields(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["client_transfer_id"] != "4fb23c36-9b31-4aec-a64d-3f80b73f5e14" || body["amount_cents"] != float64(1234) {
			t.Fatalf("body = %#v", body)
		}
		if _, exists := body["amount"]; exists {
			t.Fatalf("unexpected legacy amount field: %#v", body)
		}
	})
	_, err := c.TransferBetweenSubaccounts(context.Background(), ApplySubaccountTransferRequest{
		ClientTransferID: "4fb23c36-9b31-4aec-a64d-3f80b73f5e14",
		FromSubaccount:   0,
		ToSubaccount:     1,
		AmountCents:      1234,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestOrderMutationsSendSubaccount(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("subaccount") != "7" {
			t.Errorf("subaccount = %q", r.URL.Query().Get("subaccount"))
		}
		switch {
		case r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"order_id":"order","reduced_by":"1.00"}`))
		default:
			_, _ = w.Write([]byte(`{}`))
		}
	})
	subaccount := 7
	if _, err := c.CancelMarginOrder(context.Background(), "order", &subaccount); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AmendMarginOrder(context.Background(), "order", &subaccount, AmendMarginOrderRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.DecreaseMarginOrder(context.Background(), "order", &subaccount, DecreaseMarginOrderRequest{}); err != nil {
		t.Fatal(err)
	}
}

func TestParseErrorBody_NestedAndFlat(t *testing.T) {
	nested := parseErrorBody([]byte(`{"error":{"code":"available_balance_too_low","message":"available balance too low"}}`))
	if nested.Code != "available_balance_too_low" || nested.Message != "available balance too low" {
		t.Fatalf("nested: %+v", nested)
	}
	flat := parseErrorBody([]byte(`{"code":"x","message":"y","details":"z"}`))
	if flat.Code != "x" || flat.Message != "y" || flat.Details != "z" {
		t.Fatalf("flat: %+v", flat)
	}
}

func TestDo_NestedErrorBody(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"error":{"code":"available_balance_too_low","message":"available balance too low"}}`))
	})
	err := c.Get(context.Background(), "/margin/balance", nil, &GetMarginBalanceResponse{})
	ae, ok := err.(*APIError)
	if !ok {
		t.Fatalf("%T %v", err, err)
	}
	if ae.Body.Code != "available_balance_too_low" {
		t.Fatalf("%+v", ae.Body)
	}
	if !strings.Contains(ae.Error(), "available balance too low") {
		t.Fatal(ae.Error())
	}
}

func TestCancelMarginOrder_DecodesReducedBy(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method %s", r.Method)
		}
		_, _ = w.Write([]byte(`{"order_id":"oid","client_order_id":"cid","reduced_by":"3.00"}`))
	})
	out, err := c.CancelMarginOrder(context.Background(), "oid", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.OrderID != "oid" || out.ReducedBy != "3.00" {
		t.Fatalf("%+v", out)
	}
}

func TestCreateOrder_JSONUsesFixedPointStrings(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]any
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatal(err)
		}
		// Ensure count/price are JSON strings, not numbers.
		if _, ok := raw["count"].(string); !ok {
			t.Fatalf("count type %T", raw["count"])
		}
		if _, ok := raw["price"].(string); !ok {
			t.Fatalf("price type %T", raw["price"])
		}
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"order_id":"o","fill_count":"0.00","remaining_count":"1.00"}`))
	})
	_, err := c.CreateMarginOrder(context.Background(), CreateMarginOrderRequest{
		Ticker: "T", ClientOrderID: "c", Side: "bid", Count: "1.00", Price: "6.2000",
		TimeInForce: "good_till_canceled", SelfTradePreventionType: "maker",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestOrderGroupRequestsUseSubaccountAndDecodeDetails(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/create") {
			var req CreateOrderGroupRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Subaccount == nil || *req.Subaccount != 7 {
				t.Fatalf("request = %+v, %v", req, err)
			}
			_, _ = w.Write([]byte(`{"order_group_id":"group","subaccount":7}`))
			return
		}
		if r.URL.Query().Get("subaccount") != "7" {
			t.Errorf("subaccount = %q", r.URL.Query().Get("subaccount"))
		}
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/group") {
			_, _ = w.Write([]byte(`{"is_auto_cancel_enabled":true,"orders":["order"]}`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	})
	subaccount := 7
	if _, err := c.GetMarginOrderGroups(context.Background(), &subaccount); err != nil {
		t.Fatal(err)
	}
	if _, err := c.CreateMarginOrderGroup(context.Background(), CreateOrderGroupRequest{Subaccount: &subaccount}); err != nil {
		t.Fatal(err)
	}
	details, err := c.GetMarginOrderGroup(context.Background(), "group", &subaccount)
	if err != nil || !details.IsAutoCancelEnabled || len(details.Orders) != 1 {
		t.Fatalf("details = %+v, %v", details, err)
	}
	if err := c.DeleteMarginOrderGroup(context.Background(), "group", &subaccount); err != nil {
		t.Fatal(err)
	}
	if err := c.ResetMarginOrderGroup(context.Background(), "group", &subaccount); err != nil {
		t.Fatal(err)
	}
	if err := c.TriggerMarginOrderGroup(context.Background(), "group", &subaccount); err != nil {
		t.Fatal(err)
	}
	if err := c.UpdateMarginOrderGroupLimit(context.Background(), "group", &subaccount, UpdateOrderGroupLimitRequest{}); err != nil {
		t.Fatal(err)
	}
}

func TestGetMarginMarket_KeepsBookAndMarks(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"market":{"ticker":"X","title":"T","status":"active","contract_size":"1.000000","tick_size":"0.01","fractional_trading_enabled":true,"exchange_index":1,"bid":"1.00","ask":"1.01","settlement_mark_price":{"price":"1.005","ts_ms":9}}}`))
	})
	out, err := c.GetMarginMarket(context.Background(), "X")
	if err != nil {
		t.Fatal(err)
	}
	if out.Market.Bid != "1.00" || out.Market.Ask != "1.01" || out.Market.SettlementMarkPrice == nil || out.Market.SettlementMarkPrice.Price != "1.005" {
		t.Fatalf("%+v", out.Market)
	}
}

func TestGetMarginMarketCandlesticks_KeepsBidAskOHLC(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ticker":"X","candlesticks":[{"end_period_ts":1,"bid":{"open":"1","low":"1","high":"1","close":"1"},"ask":{"open":"2","low":"2","high":"2","close":"2"},"price":{"open":"1.5","low":"1.5","high":"1.5","close":"1.5","mean":"1.5","previous":"1.4"},"volume":"3.00","volume_notional_value_dollars":"4.00","open_interest":"5.00","open_interest_notional_value_dollars":"6.00"}]}`))
	})
	out, err := c.GetMarginMarketCandlesticks(context.Background(), "X", CandlesParams{StartTs: 1, EndTs: 2, PeriodInterval: 60})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Candlesticks) != 1 || out.Candlesticks[0].Bid.Close != "1" || out.Candlesticks[0].Ask.Close != "2" {
		t.Fatalf("%+v", out)
	}
	if out.Candlesticks[0].VolumeNotionalValueDollars != "4.00" || out.Candlesticks[0].OpenInterestNotionalValueDollars != "6.00" {
		t.Fatalf("%+v", out.Candlesticks[0])
	}
}

func TestGetMarginBalanceCanRequestAvailableBalance(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("compute_available_balance") != "true" {
			t.Errorf("compute_available_balance = %q", r.URL.Query().Get("compute_available_balance"))
		}
		_, _ = w.Write([]byte(`{"subaccount_balances":[],"settled_funds":"0"}`))
	})
	if _, err := c.GetMarginBalance(context.Background(), true); err != nil {
		t.Fatal(err)
	}
}
