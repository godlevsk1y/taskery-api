package errorsx

import "errors"

// IsAny reports whether err matches any of the target errors.
// It uses errors.Is for comparison.
func IsAny(err error, targets ...error) bool {
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}
