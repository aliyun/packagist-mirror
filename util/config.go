package util

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
	packagesJsonKey = "set:packages.json"
	packagesNoData  = "set:packages-nodata"
	distsNoMetaKey  = "set:dists-meta-missing"

	distSet          = "set:dists"
	providerSet      = "set:providers"
	packageV1Set     = "set:packagesV1"
	packageV1SetHash = "set:packagesV1-Hash"
	packageV2Set     = "set:packagesV2"
	versionsSet      = "set:versions"

	distQueue      = "queue:dists"
	distQueueRetry = "queue:dists-Retry"
	providerQueue  = "queue:providers"
	packageP1Queue = "queue:packagesV1"
	packageV2Queue = "queue:packagesV2"
)

var (
	// Wg Concurrency control
	Wg     sync.WaitGroup
	config = new(Config)
)

// Config Mirror Config
type Config struct {
	UserAgent            string `yaml:"USER_AGENT"`
	RedisAddr            string `yaml:"REDIS_ADDR"`
	RedisPassword        string `yaml:"REDIS_PASSWORD"`
	RedisDB              int    `yaml:"REDIS_DB"`
	OSSAccessKeyID       string `yaml:"OSS_ACCESS_KEY_ID"`
	OSSAccessKeySecret   string `yaml:"OSS_ACCESS_KEY_SECRET"`
	OSSEndpoint          string `yaml:"OSS_ENDPOINT"`
	OSSBucket            string `yaml:"OSS_BUCKET"`
	GithubToken          string `yaml:"GITHUB_TOKEN"`
	MirrorUrl            string `yaml:"MIRROR_URL"`
	RepoUrl              string `yaml:"REPO_URL"`
	ApiUrl               string `yaml:"API_URL"`
	ProviderUrl          string `yaml:"PROVIDER_URL"`
	DistUrl              string `yaml:"DIST_URL"`
	BuildCache           string `yaml:"BUILD_CACHE"`
	ApiIterationInterval int    `yaml:"API_ITERATION_INTERVAL"`
}

func loadConfig() {
	err := getConf(config)
	if err != nil {
		panic(err.Error())
	}
}

func getConf(conf *Config) (err error) {
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
		return
	}
	defer ymlFile.Close()
	yamlContent, err := ioutil.ReadAll(ymlFile)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(yamlContent, conf)
	if err != nil {
		return
	}

	err = validateConfig(conf)
	return
}

func validateConfig(conf *Config) (err error) {
	if conf.RedisAddr == "" {
		err = errors.New("please set necessary environment variable: REDIS_ADDR")
		return
	}

	if conf.RedisPassword == "" {
		err = errors.New("please set necessary environment variable: REDIS_PASSWORD")
		return
	}

	if conf.OSSAccessKeyID == "" {
		err = errors.New("please set necessary environment variable: OSS_ACCESS_KEY_ID")
		return
	}

	if conf.OSSAccessKeySecret == "" {
		err = errors.New("please set necessary environment variable: OSS_ACCESS_KEY_SECRET")
		return
	}

	if conf.OSSEndpoint == "" {
		err = errors.New("please set necessary environment variable: OSS_ENDPOINT")
		return
	}

	if conf.OSSBucket == "" {
		err = errors.New("please set necessary environment variable: OSS_BUCKET")
		return
	}

	if conf.GithubToken == "" {
		err = errors.New("please set necessary environment variable: GITHUB_TOKEN")
		return
	}

	if conf.MirrorUrl == "" {
		err = errors.New("please set necessary environment variable: MIRROR_URL")
		return
	}

	if conf.RepoUrl == "" {
		err = errors.New("please set necessary environment variable: REPO_URL")
		return
	}

	if conf.ApiUrl == "" {
		err = errors.New("please set necessary environment variable: API_URL")
		return
	}

	if conf.ProviderUrl == "" {
		err = errors.New("please set necessary environment variable: PROVIDER_URL")
		return
	}

	if conf.DistUrl == "" {
		err = errors.New("please set necessary environment variable: DIST_URL")
		return
	}

	if conf.BuildCache == "" {
		err = errors.New("please set necessary environment variable: BUILD_CACHE")
		return
	}

	if conf.UserAgent == "" {
		err = errors.New("please set necessary environment variable: USER_AGENT")
		return
	}

	if conf.ApiIterationInterval <= 0 {
		err = errors.New("please set necessary environment variable: API_ITERATION_INTERVAL")
		return
	}

	return
}
