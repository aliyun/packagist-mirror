package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	var config = new(Config)
	err := config.ValidateConfig()
	assert.Equal(t, "Missing configuration: REDIS_ADDR", err.Error())
	config.RedisAddr = "addr"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: REDIS_PASSWORD", err.Error())
	config.RedisPassword = "pass"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: OSS_ACCESS_KEY_ID", err.Error())
	config.OSSAccessKeyID = "oss ak id"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: OSS_ACCESS_KEY_SECRET", err.Error())
	config.OSSAccessKeySecret = "oss ak secret"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: OSS_ENDPOINT", err.Error())
	config.OSSEndpoint = "endpoint"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: OSS_BUCKET", err.Error())
	config.OSSBucket = "bucket"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: GITHUB_TOKEN", err.Error())
	config.GithubToken = "token"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: MIRROR_URL", err.Error())
	config.MirrorUrl = "mirror url"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: REPO_URL", err.Error())
	config.RepoUrl = "repo url"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: API_URL", err.Error())
	config.ApiUrl = "api url"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: PROVIDER_URL", err.Error())
	config.ProviderUrl = "provider url"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: DIST_URL", err.Error())
	config.DistUrl = "dist url"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: BUILD_CACHE", err.Error())
	config.BuildCache = "build cache"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: USER_AGENT", err.Error())
	config.UserAgent = "ua"
	err = config.ValidateConfig()
	assert.Equal(t, "Missing configuration: API_ITERATION_INTERVAL", err.Error())
	config.ApiIterationInterval = 10
	err = config.ValidateConfig()
	assert.Nil(t, err)
}
