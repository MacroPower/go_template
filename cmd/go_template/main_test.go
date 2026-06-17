package main_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	main "github.com/MacroPower/go_template/cmd/go_template"
)

func TestHello(t *testing.T) {
	t.Parallel()

	want := "Hello World!"

	sb := strings.Builder{}
	require.NoError(t, main.Hello(&sb))
	require.Equal(t, want, sb.String())
}
