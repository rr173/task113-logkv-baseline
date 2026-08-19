package ttl

import "time"

func Normalize(at time.Time) int64 {
	if at.IsZero() {
		return 0
	}
	return at.UnixNano()
}
func Remaining(expiresAt int64, now time.Time) time.Duration {
	if expiresAt <= 0 {
		return 0
	}
	return time.Until(time.Unix(0, expiresAt))
}
func IsExpired(expiresAt int64, now time.Time) bool {
	return expiresAt > 0 && !time.Unix(0, expiresAt).After(now)
}
