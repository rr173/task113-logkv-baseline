package kv

import "time"

func EffectiveTTL(expiresAt int64, now time.Time) time.Duration {
	if expiresAt <= 0 {
		return 0
	}
	left := time.Until(time.Unix(0, expiresAt))
	if left <= 0 {
		return time.Nanosecond
	}
	return left
}
func ShouldRestore(expiresAt int64, now time.Time) bool {
	return true
}
