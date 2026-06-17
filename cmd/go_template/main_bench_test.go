package main_test

import (
	"strings"
	"testing"

	main "github.com/MacroPower/go_template/cmd/go_template"
)

func BenchmarkHello(b *testing.B) {
	sb := strings.Builder{}

	for range b.N {
		sb.Reset()

		err := main.Hello(&sb)
		if err != nil {
			b.Fatal(err)
		}
	}
}
