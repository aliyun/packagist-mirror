package util

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
)

var redisClient *redis.Client

func sAdd(key string, member string) {
	redisClient.SAdd(key, member).Result()
}

func sRem(key string, member string) {
	redisClient.SRem(key, member).Result()
}

func sPop(key string) string {
	value, err := redisClient.SPop(key).Result()
	if err != nil {
		return ""
	}

	return value
}

func sIsMember(key string, member string) bool {
	exist, err := redisClient.SIsMember(key, member).Result()
	if err != nil {
		return false
	}
	return exist
}

func sCard(key string) int64 {
	num, err := redisClient.SCard(key).Result()
	if err != nil {
		num = 0
	}
	return num
}

func hDel(key string, field string) {
	redisClient.HDel(key, field)
}

func hGet(key, field string) string {
	value, err := redisClient.HGet(key, field).Result()
	if err != nil {
		//fmt.Println(key, err.Error())
	}
	return value
}

func hLen(key string) int64 {
	num, err := redisClient.HLen(key).Result()
	if err != nil {
		num = 0
	}
	return num
}

func hGetValue(key string, field string, value string) bool {
	return hGet(key, field) == value
}

func makeSucceed(key string, field string) {
	sAdd(key, field)
	removeStatusCodeFailed(key, field, 403)
	removeStatusCodeFailed(key, field, 404)
	removeStatusCodeFailed(key, field, 410)
	removeStatusCodeFailed(key, field, 500)
	removeStatusCodeFailed(key, field, 502)
	removeFailed(key, field)
}

func makeFailed(key string, field string, err error) {
	key += "-failed"
	// hSet(key, field, err.Error())
}

func removeFailed(key string, field string) {
	key += "-failed"
	hDel(key, field)
}

func countFailed(key string) int64 {
	key += "-failed"
	return hLen(key)
}

func makeStatusCodeFailed(key string, statusCode int, field string) {
	key += "-" + strconv.Itoa(statusCode)
	sAdd(key, field)
}

func removeStatusCodeFailed(key string, field string, statusCode int) {
	key += "-" + strconv.Itoa(statusCode)
	sRem(key, field)
}

func countStatusCodedFailed(key string, statusCode int) int64 {
	key += "-" + strconv.Itoa(statusCode)
	return sCard(key)
}

func getLastTimestamp() string {
	value, err := redisClient.Get("lastTimestamp").Result()
	if err != nil {
		return ""
	}

	return value
}

func setLastTimestamp(timestamp string) {
	err := redisClient.Set("lastTimestamp", timestamp, 0).Err()
	if err != nil {
		fmt.Println(timestamp, err.Error())
	}
}
