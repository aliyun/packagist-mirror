package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
