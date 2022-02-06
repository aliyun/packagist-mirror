package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func getSha256Sum(content []byte) string {
	sh := sha256.New()
	sh.Write(content)
	return hex.EncodeToString(sh.Sum(nil))
}

func GetBody(url string) (body []byte, err error) {
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

	if resp.StatusCode != 200 {
		err = fmt.Errorf("Get %s failed with code %d", url, resp.StatusCode)
		return
	}

	body, err = ioutil.ReadAll(resp.Body)
	return
}

func getTomorrow() time.Time {
	timeStr := time.Now().Format("2006-01-02")
	t2, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	return t2.AddDate(0, 0, 1)
}

func getTodayKey(key string) string {
	return key + "-" + time.Now().Format("2006-01-02")
}
