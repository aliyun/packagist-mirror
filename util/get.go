package util

import (
	"fmt"
	"net/http"
)

func mirrorUrl(path string) string {
	return config.MirrorUrl + path
}

func packagistUrl(url string) string {
	return config.DataUrl + url
}

func packagistGet(url string, processName string) (*http.Response, error) {
	return getJSON(packagistUrl(url), processName)
}

func mirrorGet(path string, processName string) (*http.Response, error) {
	return get(mirrorUrl(path), processName)
}

func getJSON(url string, processName string) (*http.Response, error) {
	fmt.Println(processName, "Get Downloading", url)
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Alibaba")
	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("Accept-Encoding", "gzip")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(processName, "Get Error", err.Error())
	} else if resp.StatusCode == 200 {
		fmt.Println(processName, "Get Downloaded", url)
	} else if resp.StatusCode != 200 {
		fmt.Println(processName, "Get Error", resp.StatusCode, url)
	}

	return resp, err
}

func get(url string, processName string) (*http.Response, error) {
	fmt.Println(processName, "Get Downloading", url)
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(processName, "Get Error", err.Error())
	} else if resp.StatusCode == 200 {
		fmt.Println(processName, "Get Downloaded", url)
	} else if resp.StatusCode != 200 {
		fmt.Println(processName, "Get Error", resp.StatusCode, url)
	}

	return resp, err
}
