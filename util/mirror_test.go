package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMirror(t *testing.T) {
	mirror := NewMirror("providerUrl", "distUrl", 5)
	assert.Equal(t, "providerUrl", mirror.providerUrl)
}
