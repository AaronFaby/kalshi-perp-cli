package cli

import "testing"

func TestValidateOrderEnums(t *testing.T) {
	if err := validateSide("buy"); err == nil {
		t.Fatal("buy")
	}
	if err := validateSide("bid"); err != nil {
		t.Fatal(err)
	}
	if err := validateTIF("gtc"); err == nil {
		t.Fatal("gtc")
	}
	if err := validateTIF("good_till_canceled"); err != nil {
		t.Fatal(err)
	}
	if err := validateSTP("fok"); err == nil {
		t.Fatal("fok")
	}
	if err := validateSTP("maker"); err != nil {
		t.Fatal(err)
	}
	if err := validateFixedPoint("price", "not-money"); err == nil {
		t.Fatal("price")
	}
	if err := validateFixedPoint("count", "1.00"); err != nil {
		t.Fatal(err)
	}
	if err := validateExchangeInstance("source", "margin"); err == nil {
		t.Fatal("margin")
	}
	if err := validateExchangeInstance("destination", "margined"); err != nil {
		t.Fatal(err)
	}
}

func TestContinueCursor(t *testing.T) {
	next, err := continueCursor(false, "", "abc", 1)
	if err != nil || next != "" {
		t.Fatalf("single page: %q %v", next, err)
	}
	next, err = continueCursor(true, "", "abc", 1)
	if err != nil || next != "abc" {
		t.Fatalf("advance: %q %v", next, err)
	}
	if _, err := continueCursor(true, "abc", "abc", 2); err == nil {
		t.Fatal("stuck cursor")
	}
	if _, err := continueCursor(true, "a", "b", maxPaginationPages); err == nil {
		t.Fatal("page cap")
	}
	next, err = continueCursor(true, "a", "", 1)
	if err != nil || next != "" {
		t.Fatalf("done: %q %v", next, err)
	}
}
