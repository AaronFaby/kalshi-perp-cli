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
		_, _ = w.Write([]byte(`{}`))
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
