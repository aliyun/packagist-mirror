package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	var config = new(Config)
	config.RedisAddr = "addr"
	config.RedisPassword = "pass"
	config.OSSAccessKeyID = "oss ak id"
	config.OSSAccessKeySecret = "oss ak secret"
	config.OSSEndpoint = "endpoint"
	config.OSSBucket = "bucket"
	config.GithubToken = "token"
	config.MirrorUrl = "mirror url"
	config.RepoUrl = "repo url"
	config.ApiUrl = "api url"
	config.ProviderUrl = "provider url"
	config.DistUrl = "dist url"
	config.BuildCache = "build cache"
	config.UserAgent = "ua"
	config.ApiIterationInterval = 10
	err := validateConfig(config)
	assert.Nil(t, err)
}
