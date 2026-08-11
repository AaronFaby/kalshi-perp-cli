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
	if err := json.Unmarshal([]byte(`{"taker_side":"ask"}`), &trade); err != nil || trade.TakerSide != "ask" {
		t.Fatalf("trade = %+v, %v", trade, err)
	}

	var fill MarginFill
	if err := json.Unmarshal([]byte(`{"entry_price":"1.20","fees":"0.01","realized_pnl":"0.30"}`), &fill); err != nil || fill.EntryPrice != "1.20" || fill.Fees != "0.01" || fill.RealizedPnL != "0.30" {
		t.Fatalf("fill = %+v, %v", fill, err)
	}
}
