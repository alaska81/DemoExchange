package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenSHA1(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{"", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},      // empty string case
		{"hello", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"}, // non-empty string case
		{"12345", "8cb2237d0679ca88db6464eac60da96345513964"}, // alphanumeric string case
	}

	for _, c := range cases {
		result := GenSHA1(c.input)
		assert.Equal(t, c.output, result)
	}
}
