package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

const (
	providerKey         = "mirror:providers"
	versionsKey         = "mirror:versions"
	packageKey          = "mirror:packages"
	providerHashFileKey = "mirror:providerHashFile"
	packageHashFileKey  = "mirror:packageHashFile"
	distsKey            = "mirror:dist"
	distsNoMetaKey      = "mirror:dists:meta:missing"
	packagesNoData      = "mirror:packages:nodata"
)

var (
	wg     sync.WaitGroup
	config = new(Config)
)

// Config Mirror Config
type Config struct {
	RedisAddr          string `yaml:"REDIS_ADDR"`
	RedisPassword      string `yaml:"REDIS_PASSWORD"`
	RedisDB            int    `yaml:"REDIS_DB"`
	OSSAccessKeyID     string `yaml:"OSS_ACCESS_KEY_ID"`
	OSSAccessKeySecret string `yaml:"OSS_ACCESS_KEY_SECRET"`
	OSSEndpoint        string `yaml:"OSS_ENDPOINT"`
	OSSBucket          string `yaml:"OSS_BUCKET"`
	GithubToken        string `yaml:"GITHUB_TOKEN"`
	MirrorUrl          string `yaml:"MIRROR_URL"`
	DataUrl            string `yaml:"DATA_URL"`
	ProviderUrl        string `yaml:"PROVIDER_URL"`
	DistUrl            string `yaml:"DIST_URL"`
}

func loadConfig() {
	err := getConf(config)
	if err != nil {
		panic(err.Error())
	}
}

func getConf(conf *Config) error {
	system := runtime.GOOS
	path, _ := os.Getwd()
	fmt.Println(path, strings.TrimRight(path, "/main"))
	if system == "windows" {
		path = strings.TrimRight(path, "\\main") + "\\packagist.yml"
	} else {
		path = strings.TrimRight(path, "/main") + "/packagist.yml"
	}
	ymlFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer ymlFile.Close()
	yamlContent, err := ioutil.ReadAll(ymlFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlContent, conf)
	if err != nil {
		return err
	}

	if conf.RedisAddr == "" {
		err = errors.New("please set necessary environment variable: REDIS_ADDR")
	}

	if conf.RedisPassword == "" {
		err = errors.New("please set necessary environment variable: REDIS_PASSWORD")
	}

	if conf.OSSAccessKeyID == "" {
		err = errors.New("please set necessary environment variable: OSS_ACCESS_KEY_ID")
	}

	if conf.OSSAccessKeySecret == "" {
		err = errors.New("please set necessary environment variable: OSS_ACCESS_KEY_SECRET")
	}

	if conf.OSSEndpoint == "" {
		err = errors.New("please set necessary environment variable: OSS_ENDPOINT")
	}

	if conf.OSSBucket == "" {
		err = errors.New("please set necessary environment variable: OSS_BUCKET")
	}

	if conf.GithubToken == "" {
		err = errors.New("please set necessary environment variable: GITHUB_TOKEN")
	}

	if conf.MirrorUrl == "" {
		err = errors.New("please set necessary environment variable: MIRROR_URL")
	}

	if conf.DataUrl == "" {
		err = errors.New("please set necessary environment variable: DATA_URL")
	}

	if conf.ProviderUrl == "" {
		err = errors.New("please set necessary environment variable: PROVIDER_URL")
	}

	if conf.DistUrl == "" {
		err = errors.New("please set necessary environment variable: DIST_URL")
	}

	return err
}
