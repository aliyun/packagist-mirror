package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCDN(t *testing.T) {
	cdn := NewCDN(false, "https://mirrors.aliyun.com/composer/")
	err := cdn.WarmUp("p2/alibabacloud/sdk.json")
	assert.Nil(t, err)
}
