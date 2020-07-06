package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/ioutil"
	"time"
)

var versions = make(map[string][]stable)
var versionsContentCache []byte

type stable struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	MinPhp  int    `json:"min-php"`
}

func composerPhar(name string, num int) {

	for {
		// Get latest stable version
		versionUrl := "https://getcomposer.org/versions"
		resp, err := get(versionUrl, getProcessName(name, num))
		if err != nil {
			continue
		}

		if resp.StatusCode != 200 {
			continue
		}

		versionsContent, err := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			fmt.Println(getProcessName(name, num), versionUrl, err.Error())
			continue
		}

		if bytes.Equal(versionsContentCache, versionsContent) {
			continue
		}

		// Sync versions
		options := []oss.Option{
			oss.ContentType("application/json"),
		}
		err = putObject(getProcessName(name, num), "versions", bytes.NewReader(versionsContent), options...)
		if err != nil {
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
		phar, err := get("https://getcomposer.org"+versions["stable"][0].Path, getProcessName(name, num))
		if err != nil {
			continue
		}

		if phar.StatusCode != 200 {
			continue
		}

		composerPhar, err := ioutil.ReadAll(phar.Body)
		_ = putObject(getProcessName(name, num), "composer.phar", bytes.NewReader(composerPhar))
		_ = putObject(getProcessName(name, num), "download/"+versions["stable"][0].Version+"/composer.phar", bytes.NewReader(composerPhar))
		_ = phar.Body.Close()

		// Like https://getcomposer.org/download/1.9.1/composer.phar.sig

		options = []oss.Option{
			oss.ContentType("application/json"),
		}

		sig, err := get("https://getcomposer.org"+versions["stable"][0].Path+".sig", getProcessName(name, num))
		if err != nil {
			continue
		}

		if sig.StatusCode != 200 {
			continue
		}

		composerPharSig, err := ioutil.ReadAll(sig.Body)
		_ = putObject(getProcessName(name, num), "composer.phar.sig", bytes.NewReader(composerPharSig), options...)
		_ = putObject(getProcessName(name, num), "download/"+versions["stable"][0].Version+"/composer.phar.sig", bytes.NewReader(composerPharSig), options...)
		_ = sig.Body.Close()

		// Sleep
		time.Sleep(6000 * time.Second)
	}
}
