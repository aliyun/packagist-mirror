package util

import (
	"fmt"
	"net/http"
)

type CDN struct {
	buildCache bool
	mirrorURL  string
}

func NewCDN(buildCache bool, mirrorURL string) (cdn *CDN) {
	return &CDN{
		buildCache: buildCache,
		mirrorURL:  mirrorURL,
	}
}

func (cdn *CDN) WarmUp(path string) (err error) {
	if cdn.buildCache == false {
		return
	}

	url := cdn.mirrorURL + path
	client := http.Client{}
	req, err := http.NewRequest("HEAD", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("Warm-Up %s failed with %d", url, resp.StatusCode)
		return
	}

	return
}
