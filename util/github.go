package util

import "net/http"

type Github struct {
	token     string
	userAgent string
}

func NewGithub(token, userAgent string) (github *Github) {
	return &Github{
		token:     token,
		userAgent: userAgent,
	}
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
	return
}
