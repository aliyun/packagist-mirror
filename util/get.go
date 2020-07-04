package util

import (
	"fmt"
	"net/http"
)

func mirrorUrl(path string) string {
	return config.MirrorUrl + path
}

func packagistUrl(url string) string {
	return config.RepoUrl + url
}

func packagistUrlApi(url string) string {
	return config.ApiUrl + url
}

func packagistGet(url string, processName string) (*http.Response, error) {
	return getJSON(packagistUrl(url), processName)
}

func packagistGetApi(url string, processName string) (*http.Response, error) {
	return getJSON(packagistUrlApi(url), processName)
}

func cdnCache(url string, name string, num int) {
	if config.BuildCache != "true" {
		return
	}
	processName := getProcessName(name, num)
	url = mirrorUrl(url)
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(processName, "Cache Error", err.Error())
		resp.Body.Close()
		return
	}

	if resp.StatusCode == 200 {
		fmt.Println(processName, "Cache Build", url)
	} else if resp.StatusCode != 200 {
		fmt.Println(processName, "Cache Error", resp.StatusCode, url)
	}

	resp.Body.Close()
}

func getJSON(url string, processName string) (*http.Response, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Alibaba")
	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("Accept-Encoding", "gzip")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(processName, "Get Error", err.Error())
	} else if resp.StatusCode != 200 {
		fmt.Println(processName, "Get Error", resp.StatusCode, url)
	}

	return resp, err
}

func get(url string, processName string) (*http.Response, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(processName, "Get Error", err.Error())
	} else if resp.StatusCode != 200 {
		fmt.Println(processName, "Get Error", resp.StatusCode, url)
	}

	return resp, err
}
