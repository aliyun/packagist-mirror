package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCDNWithFalse(t *testing.T) {
	cdn := NewCDN(false, "https://mirrors.aliyun.com/composer/")
	err := cdn.WarmUp("p2/alibabacloud/sdk.json")
	assert.Nil(t, err)
}

func TestCDNWithTrue(t *testing.T) {
	cdn := NewCDN(true, "https://mirrors.aliyun.com/composer/")
	err := cdn.WarmUp("p2/alibabacloud/sdk.json")
	assert.Nil(t, err)
}
