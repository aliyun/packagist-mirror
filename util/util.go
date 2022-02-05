package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func getProcessName(name string, num int) string {
	return "[" + getDateTime() + "]" + " " + name + ":" + strconv.Itoa(num)
}

func errHandler(err error) {
	fmt.Printf("Error: %s\n", err.Error())
	panic(err.Error())
}

// CheckHash Check Hash for File
func CheckHash(processName string, hash string, content []byte) bool {

	sh := sha256.New()
	sh.Write(content)
	sum := hex.EncodeToString(sh.Sum(nil))

	if hash != sum {
		fmt.Println(processName, "Wrong Hash", "Original:", hash, "Current:", sum)
		return false
	}

	return true
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
