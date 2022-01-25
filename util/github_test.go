package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDistFromGithub(t *testing.T) {
	resp, err := GetDistFromGithub("https://api.github.com/repos/aliyun/openapi-sdk-php/zipball/08136b7752d37fde14c3c2d6cbaabcb1dfa9c297", "", "ua")
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/zip", resp.Header.Get("Content-Type"))
	assert.Equal(t, "W/\"018eaf23eadb6330d5beec63be068be6494184937c84f7233563e9077c4c506b\"", resp.Header.Get("Etag"))
}
