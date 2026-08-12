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

func TestCandlestickUnmarshalsPerpsOHLC(t *testing.T) {
	raw := `{
		"end_period_ts": 1700000000,
		"bid": {"open":"1.00","low":"0.90","high":"1.10","close":"1.05"},
		"ask": {"open":"1.01","low":"0.91","high":"1.11","close":"1.06"},
		"price": {"open":"1.00","low":"0.90","high":"1.10","close":"1.05","mean":"1.02","previous":"0.99"},
		"volume":"10.00",
		"volume_notional_value_dollars":"100.0000",
		"open_interest":"50.00",
		"open_interest_notional_value_dollars":"500.0000"
	}`
	var c Candlestick
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatal(err)
	}
	if c.Bid.Close != "1.05" || c.Ask.Close != "1.06" {
		t.Fatalf("ohlc = %+v", c)
	}
	if c.VolumeNotionalValueDollars != "100.0000" || c.OpenInterestNotionalValueDollars != "500.0000" {
		t.Fatalf("notionals = %+v", c)
	}
	if c.Price.Previous == nil || *c.Price.Previous != "0.99" {
		t.Fatalf("price = %+v", c.Price)
	}
	out, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var remapped map[string]any
	if err := json.Unmarshal(out, &remapped); err != nil {
		t.Fatal(err)
	}
	if _, ok := remapped["yes_bid"]; ok {
		t.Fatalf("event-contract field leaked: %s", out)
	}
	if _, ok := remapped["bid"]; !ok {
		t.Fatalf("missing bid: %s", out)
	}
}

func TestCandlestickIgnoresEventContractFieldNames(t *testing.T) {
	var c Candlestick
	if err := json.Unmarshal([]byte(`{"end_period_ts":1,"yes_bid":{"close":"9"},"yes_ask":{"close":"8"},"bid":{"open":"1","low":"1","high":"1","close":"2"}}`), &c); err != nil {
		t.Fatal(err)
	}
	if c.Bid.Close != "2" {
		t.Fatalf("bid close = %q", c.Bid.Close)
	}
}

func TestMarginMarketKeepsBookMarksAndVolume(t *testing.T) {
	raw := `{
		"ticker":"KXBTCPERP1","title":"BTC","status":"active",
		"contract_size":"1.000000","tick_size":"0.0001",
		"fractional_trading_enabled":true,"exchange_index":1,
		"bid":"6.1900","ask":"6.2000","price":"6.1950",
		"volume":"100.00","volume_24h":"10.00",
		"volume_notional_value_dollars":"620.0000",
		"open_interest":"50.00","open_interest_notional_value_dollars":"310.0000",
		"settlement_mark_price":{"price":"6.1940","ts_ms":1},
		"liquidation_mark_price":{"price":"6.1930","ts_ms":2},
		"reference_price":{"price":"62000.00","ts_ms":3},
		"leverage_estimate":5.5
	}`
	var m MarginMarket
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	if m.Bid != "6.1900" || m.Ask != "6.2000" || m.Volume24h != "10.00" {
		t.Fatalf("book/volume = %+v", m)
	}
	if m.SettlementMarkPrice == nil || m.SettlementMarkPrice.Price != "6.1940" {
		t.Fatalf("settlement = %+v", m.SettlementMarkPrice)
	}
	if m.LiquidationMarkPrice == nil || m.ReferencePrice == nil || m.LeverageEstimate == nil || *m.LeverageEstimate != 5.5 {
		t.Fatalf("marks/leverage = %+v", m)
	}
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var remapped map[string]any
	if err := json.Unmarshal(out, &remapped); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"bid", "ask", "settlement_mark_price", "liquidation_mark_price", "reference_price", "volume_24h"} {
		if _, ok := remapped[k]; !ok {
			t.Errorf("missing %s in %s", k, out)
		}
	}
}
