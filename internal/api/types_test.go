package api

import (
	"encoding/json"
	"testing"
)

func TestResponseFieldNamesMatchSchema(t *testing.T) {
	var subaccount CreateSubaccountResponse
	if err := json.Unmarshal([]byte(`{"subaccount_number":7}`), &subaccount); err != nil || subaccount.SubaccountNumber != 7 {
		t.Fatalf("subaccount = %+v, %v", subaccount, err)
	}

	var trade MarginTrade
	if err := json.Unmarshal([]byte(`{"taker_side":"ask","trade_id":"t","ticker":"X","count":"1.00","price":"1.00"}`), &trade); err != nil || trade.TakerSide != "ask" {
		t.Fatalf("trade = %+v, %v", trade, err)
	}

	var fill MarginFill
	if err := json.Unmarshal([]byte(`{"fill_id":"f","order_id":"o","ticker":"X","side":"bid","count":"1.00","price":"1.00","entry_price":"1.20","fees":"0.01","realized_pnl":"0.30","is_taker":true}`), &fill); err != nil || fill.EntryPrice != "1.20" || fill.Fees != "0.01" || fill.RealizedPnL != "0.30" {
		t.Fatalf("fill = %+v, %v", fill, err)
	}

	var order MarginOrder
	payload := `{"order_id":"o","user_id":"u","client_order_id":"c","ticker":"X","side":"bid","price":"6.20","fill_count":"1.00","remaining_count":"0.00","last_update_reason":"fill","order_source":"user"}`
	if err := json.Unmarshal([]byte(payload), &order); err != nil || order.OrderSource != "user" || order.FillCount != "1.00" {
		t.Fatalf("order = %+v, %v", order, err)
	}

	var group GetOrderGroupResponse
	if err := json.Unmarshal([]byte(`{"is_auto_cancel_enabled":true,"orders":["a","b"],"contracts_limit_fp":"10.00"}`), &group); err != nil || !group.IsAutoCancelEnabled || len(group.Orders) != 2 {
		t.Fatalf("group = %+v, %v", group, err)
	}

	var bal GetMarginBalanceResponse
	if err := json.Unmarshal([]byte(`{"settled_funds":"45.0000","subaccount_balances":[{"subaccount":0,"position_value":"0","account_equity":"25","maintenance_margin":"0","initial_margin":"0","resting_orders_margin":"0","available_balance":"25"}]}`), &bal); err != nil || bal.SettledFunds != "45.0000" || bal.SubaccountBalances[0].AvailableBalance != "25" {
		t.Fatalf("bal = %+v, %v", bal, err)
	}

	var pos MarginPosition
	if err := json.Unmarshal([]byte(`{"subaccount":0,"market_ticker":"KXBTCPERP1","position":"11.00","entry_price":"6.35","unrealized_pnl":"0.05","fees":"0.5","is_portfolio":false}`), &pos); err != nil || pos.MarketTicker != "KXBTCPERP1" || pos.Position != "11.00" {
		t.Fatalf("pos = %+v, %v", pos, err)
	}
}

func TestRequestBodiesMarshalRequiredFields(t *testing.T) {
	order, err := json.Marshal(CreateMarginOrderRequest{
		Ticker: "T", ClientOrderID: "c", Side: "bid", Count: "1.00", Price: "6.2000",
		TimeInForce: "good_till_canceled", SelfTradePreventionType: "maker",
	})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(order, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"ticker", "client_order_id", "side", "count", "price", "time_in_force", "self_trade_prevention_type"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing %s in %s", k, order)
		}
	}

	xfer, err := json.Marshal(ApplySubaccountTransferRequest{
		ClientTransferID: "4fb23c36-9b31-4aec-a64d-3f80b73f5e14",
		FromSubaccount:   0,
		ToSubaccount:     1,
		AmountCents:      100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(xfer, &m); err != nil {
		t.Fatal(err)
	}
	if m["amount_cents"] != float64(100) {
		t.Fatalf("%v", m)
	}
	if _, ok := m["amount"]; ok {
		t.Fatal("legacy amount must not be present")
	}
}
