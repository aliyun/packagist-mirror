package util

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var versions = make(map[string][]stable)
var versionsContentCache []byte

type stable struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	MinPhp  int    `json:"min-php"`
}

func (ctx *Context) SyncComposerPhar(processName string) {

	for {
		// Each cycle requires a time slot
		time.Sleep(6000 * time.Second)

		// Get latest stable version
		versionsContent, err := GetBody("https://getcomposer.org/versions")
		if err != nil {
			// TODO: logger the error, but ignore it
			continue
		}

		if bytes.Equal(versionsContentCache, versionsContent) {
			continue
		}

		// Sync versions
		options := []oss.Option{
			oss.ContentType("application/json"),
		}

		err = ctx.ossBucket.PutObject("versions", bytes.NewReader(versionsContent), options...)
		if err != nil {
			// TODO: logger the error, but ignore it
			continue
		}

		// The cache is updated only if the push is successful
		versionsContentCache = versionsContent

		// JSON Decode
		err = json.Unmarshal(versionsContent, &versions)
		if err != nil {
			errHandler(err)
			continue
		}

		// Like https://getcomposer.org/download/1.9.1/composer.phar
		composerPhar, err := GetBody("https://getcomposer.org" + versions["stable"][0].Path)
		if err != nil {
			// TODO: logger the error, but ignore it
			continue
		}

		_ = ctx.ossBucket.PutObject("composer.phar", bytes.NewReader(composerPhar))
		_ = ctx.ossBucket.PutObject("download/"+versions["stable"][0].Version+"/composer.phar", bytes.NewReader(composerPhar))

		// Like https://getcomposer.org/download/1.9.1/composer.phar.sig
		composerPharSig, err := GetBody("https://getcomposer.org" + versions["stable"][0].Path + ".sig")
		if err != nil {
			// TODO: logger the error, but ignore it
			continue
		}

		options = []oss.Option{
			oss.ContentType("application/json"),
		}
		_ = ctx.ossBucket.PutObject("composer.phar.sig", bytes.NewReader(composerPharSig), options...)
		_ = ctx.ossBucket.PutObject("download/"+versions["stable"][0].Version+"/composer.phar.sig", bytes.NewReader(composerPharSig), options...)
	}
}
