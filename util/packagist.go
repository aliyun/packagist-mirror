package util

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Packagist struct {
	repoUrl string
	apiUrl  string
}

func NewPackagist(repoUrl string, apiUrl string) (packagist *Packagist) {
	return &Packagist{
		repoUrl: repoUrl,
		apiUrl:  apiUrl,
	}
}

type Hashes struct {
	SHA256 string `json:"sha256"`
}

type Packages struct {
	NotifyBatch      string            `json:"notify-batch"`
	ProviderIncludes map[string]Hashes `json:"provider-includes"`
}

func (packagist *Packagist) GetPackagesJSON() (content []byte, err error) {
	url := packagist.repoUrl + "packages.json"
	content, err = GetBody(url)
	return
}

func (packagist *Packagist) GetMetadataChanges(lastTimestamp string) (changes Changes, err error) {
	url := packagist.apiUrl + "metadata/changes.json?since=" + lastTimestamp
	content, err := GetBody(url)
	if err != nil {
		return
	}

	// JSON Decode
	err = json.Unmarshal(content, &changes)
	return
}

func (packagist *Packagist) GetInitTimestamp() (timestamp string, err error) {
	url := packagist.apiUrl + "metadata/changes.json"
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// req.Header.Add("User-Agent", config.UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// JSON Decode
	changesJson := make(map[string]interface{})
	err = json.Unmarshal(body, &changesJson)
	if err != nil {
		return
	}

	timestamp = strconv.FormatInt(int64(changesJson["timestamp"].(float64)), 10)
	return
}

func (packagist *Packagist) GetAllPackages() (content []byte, err error) {
	url := packagist.apiUrl + "packages/list.json"
	content, err = GetBody(url)
	return
}

func (packagist *Packagist) GetPackage(packageName string) (content []byte, err error) {
	url := packagist.apiUrl + "p2/" + packageName + ".json"
	content, err = GetBody(url)
	return
}
