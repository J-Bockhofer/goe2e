package goe2e_test

import (
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e/pkg"

	"github.com/stretchr/testify/require"
)

func TestEmptyModifierListsAreNoOps(t *testing.T) {
	rh, err := goe2e.NewRequestHandler()
	require.NoError(t, err)

	require.NoError(t, rh.ModifyRequest())
	require.NoError(t, rh.ModifyResponseBody())
	require.NoError(t, rh.ModifyResponse())
}
