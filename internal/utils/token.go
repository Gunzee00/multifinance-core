package utils

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("invalid token")
var ErrExpiredToken = errors.New("expired token")

func ParseToken(token string, ttl time.Duration) (string, time.Time, error) {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return "", time.Time{}, ErrInvalidToken
	}
	ts, err := time.Parse(time.RFC3339, parts[1])
	if err != nil {
		return "", time.Time{}, ErrInvalidToken
	}
	if time.Since(ts) > ttl {
		return "", time.Time{}, ErrExpiredToken
	}
	return parts[0], ts, nil
}
