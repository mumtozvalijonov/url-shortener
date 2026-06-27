package generator_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/mumtozvalijonov/url-shortener/internal/adapters/generator"
	"github.com/stretchr/testify/require"
)

func TestRandomGenerator_Generate(t *testing.T) {
	size := gofakeit.Uint8()%5 + 5
	generator := generator.NewRandomGenerator(size)
	got := generator.Generate()
	require.Len(t, got, int(size))
}
