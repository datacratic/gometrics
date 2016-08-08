package defaults

import (
	"net/http"
	"time"
)

// String is a simple helper to provide a fallback string when empty.
func String(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

func Int64(value, fallback int64) int64 {
	if value == 0 {
		return fallback
	}

	return value
}

// Duration is a simple helper to provide a fallback duration when zero.
func Duration(value, fallback time.Duration) time.Duration {
	if value == 0 {
		return fallback
	}

	return value
}

func Client(value *http.Client) *http.Client {
	if value == nil {
		return http.DefaultClient
	}

	return value
}
