package util

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type Github struct {
	token     string
	userAgent string
}

const Github404 = GithubError("github: 404")

type GithubError string

func (e GithubError) Error() string { return string(e) }

func NewGithub(token, userAgent string) (github *Github) {
	return &Github{
		token:     token,
		userAgent: userAgent,
	}
}

func (github *Github) Test() (err error) {
	// Create a new request using http
	req, err := http.NewRequest("GET", "https://api.github.com/zen", nil)
	if err != nil {
		return
	}

	// add authorization header to the req
	if github.token != "" {
		req.Header.Add("Authorization", "token "+github.token)
	}

	req.Header.Add("User-Agent", github.userAgent)
	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("test github failed with code %d", resp.StatusCode)
		return
	}

	_, err = ioutil.ReadAll(resp.Body)
	return
}

func (github *Github) GetDist(url string) (resp *http.Response, err error) {
	// Get dist
	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// add authorization header to the req
	if github.token != "" {
		req.Header.Add("Authorization", "token "+github.token)
	}
	req.Header.Add("User-Agent", github.userAgent)
	// Send req using http Client
	client := &http.Client{}
	resp, err = client.Do(req)

	if err != nil {
		return
	}

	if resp.StatusCode == 404 {
		err = Github404
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("request github dist(%s) failed with code %d", url, resp.StatusCode)
		return
	}

	return
}
