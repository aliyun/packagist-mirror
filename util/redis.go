package util

import (
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

var redisClient *redis.Client

func initRedisClient() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})
}

func countAll(key, member string) {
	sAdd(key, member)
}

func countToday(key, member string) {
	sAdd(getTodayKey(key), member)
	redisClient.ExpireAt(getTodayKey(key), getTomorrow())
}

func getTomorrow() time.Time {
	timeStr := time.Now().Format("2006-01-02")
	t2, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	return t2.AddDate(0, 0, 1)
}

func getTodayKey(key string) string {
	return key + "-" + time.Now().Format("2006-01-02")
}

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

func hSet(key string, field string, content string) {
	redisClient.HSet(key, field, content)
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
	hSet(key, field, err.Error())
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
