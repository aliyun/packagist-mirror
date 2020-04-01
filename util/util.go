package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

func getProcessName(name string, num int) string {
	return "[" + getDateTime() + "]" + " " + name + ":" + strconv.Itoa(num)
}

func errHandler(err error) {
	fmt.Printf("Error: %s\n", err.Error())
	panic(err.Error())
}

// GetHashFromPath Get Hash from File path
func GetHashFromPath(path string) string {
	str := strings.Split(path, "$")
	hash := strings.TrimSuffix(str[1], ".json")
	return hash
}

// CheckHash Check Hash for File
func CheckHash(process string, path string, content []byte) bool {
	hash := GetHashFromPath(path)

	fmt.Println(process, "Original Hash", hash)

	sh := sha256.New()
	sh.Write(content)
	sum := hex.EncodeToString(sh.Sum(nil))

	if hash != sum {
		fmt.Println(process, "Wrong Hash", path, sum)
	}

	return hash == sum
}
