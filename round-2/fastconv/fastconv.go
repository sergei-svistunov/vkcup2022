package fastconv

import (
	"errors"
	"math"
)

// Fast, but not the fastest :) Just no memory allocations.

const digits = "0123456789"

func Itoa(n int64, dst []byte) []byte {
	dst = dst[:cap(dst)]

	if n == math.MinInt64 {
		copy(dst, "-9223372036854775808")
		return dst[:20]
	}

	signed := n < 0
	if n < 0 {
		n *= -1
	}

	pos := 0
	for {
		dst[pos] = digits[n%10]
		pos++
		n /= 10
		if n == 0 {
			break
		}
	}

	if signed {
		dst[pos] = '-'
		pos++
	}

	for i := 0; i < pos/2; i++ {
		dst[i], dst[pos-i-1] = dst[pos-i-1], dst[i]
	}

	return dst[:pos]
}

func Atoi(b []byte) (int64, error) {
	if len(b) == 0 {
		return 0, errors.New("invalid number")
	}

	res := int64(0)
	multiplier := int64(1)
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] >= '0' && b[i] <= '9' {
			res += int64(b[i]-'0') * multiplier
		} else if i == 0 && b[0] == '-' {
			res *= -1
		} else {
			return 0, errors.New("invalid number")
		}
		multiplier *= 10
	}

	return res, nil
}
