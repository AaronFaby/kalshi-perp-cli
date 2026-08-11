package cli

import "testing"

func TestDollarsToCenticents(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"1", 10000},
		{"1.00", 10000},
		{"0.01", 100},
		{"12.3456", 123456},
		{"0", 0},
	}
	for _, tc := range cases {
		got, err := dollarsToCenticents(tc.in)
		if err != nil {
			t.Fatalf("%s: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("%s: got %d want %d", tc.in, got, tc.want)
		}
	}
}

func TestDollarsToUnitsRejectsInexactAndOverflowingAmounts(t *testing.T) {
	for _, in := range []string{"0.00001", "-1", "not-money", "1000000000000000"} {
		if _, err := dollarsToCenticents(in); err == nil {
			t.Errorf("%q: expected error", in)
		}
	}
	if got, err := dollarsToUnits("12.34", 100); err != nil || got != 1234 {
		t.Fatalf("got %d, %v", got, err)
	}
}
