package generator

import "math/rand"

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type RandomGenerator struct {
	size uint8
}

func NewRandomGenerator(size uint8) *RandomGenerator {
	return &RandomGenerator{size: size}
}

func (g RandomGenerator) Generate() string {
	b := make([]rune, g.size)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
