package utils

import "testing"

func TestFormatValue(t *testing.T) {
	cases := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"string passes through", "100.00", "100.00"},
		{"empty string", "", ""},
		{"nil becomes empty", nil, ""},
		{"int", 42, "42"},
		{"negative int", -7, "-7"},
		{"int64", int64(9000000000), "9000000000"},
		{"uint", uint(42), "42"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"float with decimals", 100.5, "100.5"},
		{"whole float has no trailing zeros", 100.0, "100"},
		{"large float is not scientific", 1000000.0, "1000000"},
		{"float32", float32(1.5), "1.5"},
		{"small float", 0.05, "0.05"},
		{"string slice falls back to %v", []string{"a", "b"}, "[a b]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatValue(tc.input)
			if got != tc.want {
				t.Errorf("FormatValue(%#v) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
