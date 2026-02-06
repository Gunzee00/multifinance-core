package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseToken_Valid(t *testing.T) {
	user := "alice"
	ts := time.Now().UTC()
	token := user + ":" + ts.Format(time.RFC3339)

	gotUser, gotTs, err := ParseToken(token, time.Hour)
	require.NoError(t, err)
	require.Equal(t, user, gotUser)
	require.WithinDuration(t, ts, gotTs, time.Second)
}

func TestParseToken_InvalidFormat(t *testing.T) {
	_, _, err := ParseToken("invalidtoken", time.Hour)
	require.Error(t, err)
	require.Equal(t, ErrInvalidToken, err)
}

func TestParseToken_BadTimestamp(t *testing.T) {
	token := "bob:badtime"
	_, _, err := ParseToken(token, time.Hour)
	require.Error(t, err)
	require.Equal(t, ErrInvalidToken, err)
}

func TestParseToken_Expired(t *testing.T) {
	user := "eve"
	ts := time.Now().Add(-48 * time.Hour).UTC()
	token := user + ":" + ts.Format(time.RFC3339)

	_, _, err := ParseToken(token, 24*time.Hour)
	require.Error(t, err)
	require.Equal(t, ErrExpiredToken, err)
}
