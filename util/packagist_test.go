package util

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetPackagesJSON(t *testing.T) {
	p := NewPackagist("https://packagist.org/", "https://packagist.org/")
	body, err := p.GetPackagesJSON()
	assert.Nil(t, err)
	pkg, err := getPackages(body)
	assert.Nil(t, err)
	assert.Equal(t, "https://packagist.org/downloads/", pkg.NotifyBatch)
	assert.Greater(t, len(pkg.ProviderIncludes), 0)
}

func TestGetInitTimestamp(t *testing.T) {
	p := NewPackagist("https://packagist.org/", "https://packagist.org/")
	timestamp, err := p.GetInitTimestamp()
	assert.Nil(t, err)
	assert.NotNil(t, timestamp)
}

func TestGetChanges(t *testing.T) {
	p := NewPackagist("https://packagist.org/", "https://packagist.org/")
	changes, err := p.GetMetadataChanges(strconv.FormatInt(time.Now().UnixMilli()*10-2000, 10))
	assert.Nil(t, err)
	assert.NotNil(t, changes.Timestamp)
}

// func TestGetAllPackages(t *testing.T) {
// 	p := NewPackagist("https://packagist.org/", "https://packagist.org/")
// 	content, err := p.GetAllPackages()
// 	assert.Nil(t, err)
// 	assert.NotNil(t, content)
// }
