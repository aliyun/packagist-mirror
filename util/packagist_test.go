package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetPackagesJSON(t *testing.T) {
	p := NewPackagist("https://packagist.org/", "https://packagist.org/", "ua")
	body, lastModified, err := p.GetPackagesJSON()
	assert.Nil(t, err)
	assert.NotNil(t, lastModified)
	pkg, err := getPackages(body)
	assert.Nil(t, err)
	assert.Equal(t, "https://packagist.org/downloads/", pkg.NotifyBatch)
	assert.Greater(t, len(pkg.ProviderIncludes), 0)
}

func TestGetInitTimestamp(t *testing.T) {
	p := NewPackagist("https://packagist.org/", "https://packagist.org/", "ua")
	timestamp, err := p.GetInitTimestamp()
	assert.Nil(t, err)
	assert.NotNil(t, timestamp)
}

func TestGetMetadataChanges(t *testing.T) {
	p := NewPackagist("https://packagist.org/", "https://packagist.org/", "ua")
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

func TestGetEmptyPackages(t *testing.T) {
	p := NewPackagist("https://packagist.org/", "https://packagist.org/", "ua")
	content, err := p.Get("p/keycdn/optimus-api$454eea274ef715bed0aef364fec823e04318850d28eb04f1c16a03a5127c071c.json")
	assert.Nil(t, err)
	response := new(Response)
	// fmt.Println(string(content))
	err = json.Unmarshal(content, &response)
	assert.NotNil(t, err)
}
func TestGet(t *testing.T) {
	p := NewPackagist("https://packagist.org/", "https://packagist.org/", "ua")
	content, err := p.Get("p/tugmaks/russian-text-uniqifier$6097fb12f9723ce4b0fcf4dc08daf998adc2005af411fdc793aec5bfa7ce449f.json")
	assert.Nil(t, err)
	response := new(Response)
	fmt.Println(string(content))
	err = json.Unmarshal(content, &response)
}
