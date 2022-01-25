package util

import "net/http"

func GetDistFromGithub(url string, token string, userAgent string) (resp *http.Response, err error) {
	// Get dist
	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// add authorization header to the req
	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}
	req.Header.Add("User-Agent", userAgent)
	// Send req using http Client
	client := &http.Client{}
	resp, err = client.Do(req)
	return
}
