package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPackagesJSON(t *testing.T) {
	p := NewPackagist("https://packagist.org/", "https://packagist.org/")
	body, err := p.GetPackagesJSON()
	assert.Nil(t, err)
	pkg, err := getPackages(body)
	assert.Nil(t, err)
	assert.Equal(t, "https://packagist.org/downloads/", pkg.NotifyBatch)
}

func TestGetInitTimestamp(t *testing.T) {
	p := NewPackagist("https://packagist.org/", "https://packagist.org/")
	timestamp, err := p.GetInitTimestamp()
	assert.Nil(t, err)
	assert.NotNil(t, timestamp)
}

// func TestGetAllPackages(t *testing.T) {
// 	p := NewPackagist("https://packagist.org/", "https://packagist.org/")
// 	content, err := p.GetAllPackages()
// 	assert.Nil(t, err)
// 	assert.NotNil(t, content)
// }
