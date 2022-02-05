package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackagist(t *testing.T) {
	p := NewPackagist("https://packagist.org/", "")
	body, err := p.GetPackagesJSON()
	assert.Nil(t, err)
	fmt.Println(string(body))
	assert.NotNil(t, body)
}
