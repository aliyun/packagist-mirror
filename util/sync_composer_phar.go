package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-redis/redis"
)

type Stable struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	MinPhp  int    `json:"min-php"`
}

func (ctx *Context) SyncComposerPhar(processName string) {
	var logger = NewLogger(processName)

	logger.Info("init sync composer.phar")
	for {
		err := syncComposerPhar(ctx, logger)
		if err != nil {
			logger.Error("Sync composer.phar failed: " + err.Error())
		}
		// Each cycle requires a time slot
		time.Sleep(6000 * time.Second)
	}
}

func syncComposerPhar(ctx *Context, logger *MyLogger) (err error) {
	// Get latest stable version
	versionsContent, err := GetBody("https://getcomposer.org/versions")
	if err != nil {
		// logger the error, but ignore it
		err = fmt.Errorf("get composer versions failed: " + err.Error())
		return
	}

	var versions = make(map[string][]Stable)
	// JSON Decode
	err = json.Unmarshal(versionsContent, &versions)
	if err != nil {
		err = fmt.Errorf("unmarshal versions failed: " + err.Error())
		return
	}

	stable := versions["stable"][0]

	localStableVersion, err := ctx.redis.Get(localStableComposerVersion).Result()
	if err != nil && err != redis.Nil {
		err = fmt.Errorf("call redis failed: " + err.Error())
		return
	}

	if localStableVersion == stable.Version {
		logger.Info("The remote version is equals with local version, no need to anything")
		return
	}

	// about 2.4MB
	logger.Info("get composer.phar now")
	// Like https://getcomposer.org/download/1.9.1/composer.phar
	composerPhar, err := GetBody("https://getcomposer.org" + stable.Path)
	if err != nil {
		// logger the error, but ignore it
		err = fmt.Errorf("get composer phar failed: " + err.Error())
		return
	}

	// Like https://getcomposer.org/download/1.9.1/composer.phar.sig
	composerPharSig, err := GetBody("https://getcomposer.org" + stable.Path + ".sig")
	if err != nil {
		// logger the error, but ignore it
		err = fmt.Errorf("get stable composer.phar.sig failed: " + err.Error())
		return
	}

	// Sync versions
	options := []oss.Option{
		oss.ContentType("application/json"),
	}

	err = ctx.ossBucket.PutObject("versions", bytes.NewReader(versionsContent), options...)
	if err != nil {
		// logger the error, but ignore it
		err = fmt.Errorf("put versions to OSS failed: " + err.Error())
		return
	}

	logger.Info("put composer.phar on OSS")
	err = ctx.ossBucket.PutObject("composer.phar", bytes.NewReader(composerPhar))
	if err != nil {
		// logger the error, but ignore it
		err = fmt.Errorf("put composer.phar failed: " + err.Error())
		return
	}

	err = ctx.ossBucket.PutObject("download/"+stable.Version+"/composer.phar", bytes.NewReader(composerPhar))
	if err != nil {
		// logger the error, but ignore it
		err = fmt.Errorf("put stable composer.phar failed: " + err.Error())
		return
	}

	options = []oss.Option{
		oss.ContentType("application/json"),
	}
	err = ctx.ossBucket.PutObject("composer.phar.sig", bytes.NewReader(composerPharSig), options...)
	if err != nil {
		err = fmt.Errorf("put stable composer.phar.sig failed: " + err.Error())
		return
	}

	err = ctx.ossBucket.PutObject("download/"+stable.Version+"/composer.phar.sig", bytes.NewReader(composerPharSig), options...)
	if err != nil {
		err = fmt.Errorf("put stable(version) composer.phar.sig failed: " + err.Error())
		return
	}

	// The cache is updated only if the push is successful
	logger.Info("save stable composer version into local store")
	err = ctx.redis.Set(localStableComposerVersion, stable.Version, 0).Err()
	if err != nil {
		err = fmt.Errorf("save stable composer version failed: " + err.Error())
		return
	}

	return
}
