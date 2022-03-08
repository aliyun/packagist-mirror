package util

import (
	"errors"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
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

func LoadConfig(configPath string) (conf *Config, err error) {
	content, err := getYamlContent(configPath)
	if err != nil {
		return
	}
	conf = new(Config)
	err = yaml.Unmarshal(content, conf)
	if err != nil {
		return
	}
	err = conf.ValidateConfig()
	return
}

func getYamlContent(yamlPath string) (content []byte, err error) {
	ymlFile, err := os.Open(yamlPath)
	if err != nil {
		return
	}
	defer ymlFile.Close()
	content, err = ioutil.ReadAll(ymlFile)
	return
}

func (config *Config) ValidateConfig() (err error) {
	if config.RedisAddr == "" {
		err = errors.New("missing configuration: REDIS_ADDR")
		return
	}

	if config.RedisPassword == "" {
		err = errors.New("missing configuration: REDIS_PASSWORD")
		return
	}

	if config.OSSAccessKeyID == "" {
		err = errors.New("missing configuration: OSS_ACCESS_KEY_ID")
		return
	}

	if config.OSSAccessKeySecret == "" {
		err = errors.New("missing configuration: OSS_ACCESS_KEY_SECRET")
		return
	}

	if config.OSSEndpoint == "" {
		err = errors.New("missing configuration: OSS_ENDPOINT")
		return
	}

	if config.OSSBucket == "" {
		err = errors.New("missing configuration: OSS_BUCKET")
		return
	}

	if config.GithubToken == "" {
		err = errors.New("missing configuration: GITHUB_TOKEN")
		return
	}

	if config.MirrorUrl == "" {
		err = errors.New("missing configuration: MIRROR_URL")
		return
	}

	if config.RepoUrl == "" {
		err = errors.New("missing configuration: REPO_URL")
		return
	}

	if config.ApiUrl == "" {
		err = errors.New("missing configuration: API_URL")
		return
	}

	if config.ProviderUrl == "" {
		err = errors.New("missing configuration: PROVIDER_URL")
		return
	}

	if config.DistUrl == "" {
		err = errors.New("missing configuration: DIST_URL")
		return
	}

	if config.BuildCache == "" {
		err = errors.New("missing configuration: BUILD_CACHE")
		return
	}

	if config.UserAgent == "" {
		err = errors.New("missing configuration: USER_AGENT")
		return
	}

	if config.ApiIterationInterval <= 0 {
		err = errors.New("missing configuration: API_ITERATION_INTERVAL")
		return
	}

	return
}
