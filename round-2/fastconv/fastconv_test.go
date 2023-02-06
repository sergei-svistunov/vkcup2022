package fastconv

import (
	"fmt"
	"math"
	"testing"
)

func TestItoa(t *testing.T) {
	t.Parallel()
	buf := [20]byte{}
	for _, testCase := range []struct {
		input    int64
		expected string
	}{
		{12345, "12345"},
		{1, "1"},
		{0, "0"},
		{-1, "-1"},
		{-12345, "-12345"},
		{math.MinInt64, "-9223372036854775808"},
		{math.MinInt64 + 1, "-9223372036854775807"},
		{math.MaxInt64, "9223372036854775807"},
	} {
		t.Run(fmt.Sprintf(`From %d`, testCase.input), func(t *testing.T) {
			t.Parallel()
			if got := Itoa(testCase.input, buf[:]); string(got) != testCase.expected {
				t.Fatalf("Invalid convertion from %d: expected %s, got %s", testCase.input, testCase.expected, string(got))
			}
		})
	}
}

func TestAtoi(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		expected int64
		input    string
		wuthErr  bool
	}{
		{12345, "12345", false},
		{1, "1", false},
		{0, "0", false},
		{-1, "-1", false},
		{-12345, "-12345", false},
		{math.MinInt64, "-9223372036854775808", false},
		{math.MinInt64 + 1, "-9223372036854775807", false},
		{math.MaxInt64, "9223372036854775807", false},
		{0, "abc", true},
		{0, "123abc", true},
		{0, "", true},
	} {
		t.Run(fmt.Sprintf(`From "%s"`, testCase.input), func(t *testing.T) {
			got, err := Atoi([]byte(testCase.input))
			if err != nil && !testCase.wuthErr {
				t.Parallel()
				t.Fatal(err)
			}

			if testCase.wuthErr && err == nil {
				t.Fatalf("Expected but not received an error for '%s'", testCase.input)
			}

			if got != testCase.expected {
				t.Fatalf("Invalid convertion from %s: expected %d, got %d", testCase.input, testCase.expected, got)
			}
		})
	}
}
