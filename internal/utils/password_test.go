package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashAndComparePassword(t *testing.T) {
	plain := "secret123"

	hash, err := HashPassword(plain)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	// correct password
	err = ComparePassword(hash, plain)
	require.NoError(t, err)

	// wrong password
	err = ComparePassword(hash, "wrong")
	require.Error(t, err)
}

func TestHashEmptyPassword(t *testing.T) {
	plain := ""

	hash, err := HashPassword(plain)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	err = ComparePassword(hash, plain)
	require.NoError(t, err)
}
