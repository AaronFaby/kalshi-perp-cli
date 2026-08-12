package cli

import (
	"fmt"
	"math/big"
	"strings"
)

func validateSide(s string) error {
	switch s {
	case "bid", "ask":
		return nil
	default:
		return fmt.Errorf("--side must be bid or ask")
	}
}

func validateTIF(s string) error {
	switch s {
	case "fill_or_kill", "good_till_canceled", "immediate_or_cancel":
		return nil
	default:
		return fmt.Errorf("--tif must be fill_or_kill, good_till_canceled, or immediate_or_cancel")
	}
}

func validateSTP(s string) error {
	switch s {
	case "taker_at_cross", "maker":
		return nil
	default:
		return fmt.Errorf("--stp must be taker_at_cross or maker")
	}
}

func validateFixedPoint(name, s string) error {
	if _, ok := new(big.Rat).SetString(strings.TrimSpace(s)); !ok {
		return fmt.Errorf("--%s must be a decimal string", name)
	}
	return nil
}

func validateExchangeInstance(name, s string) error {
	switch s {
	case "event_contract", "margined":
		return nil
	default:
		return fmt.Errorf("--%s must be event_contract or margined", name)
	}
}

func channelRequiresMarket(ch string) bool {
	return ch == "orderbook_delta"
}
