package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSha256Sum(t *testing.T) {
	assert.Equal(t, "64ec88ca00b268e5ba1a35678a1b5316d212f4f366b2477232534a8aeca37f3c", getSha256Sum([]byte("Hello world")))
}
