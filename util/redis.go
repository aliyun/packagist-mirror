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

func count(key, field string) {
	redisClient.HSet(key, field, getDateTime())
	redisClient.HSet(getTodayKey(key), field, getDateTime())
	redisClient.ExpireAt(getTodayKey(key), getTomorrow())
}

func getTomorrow() time.Time {
	timeStr := time.Now().Format("2006-01-02")
	t2, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	return t2.AddDate(0, 0, 1)
}

func getTodayKey(key string) string {
	return key + ":" + time.Now().Format("2006-01-02")
}

func hExists(key, field string) bool {
	exist, err := redisClient.HExists(key, field).Result()
	if err != nil {
		fmt.Println(err.Error())
	}
	if exist {
		return true
	}
	return false
}

func isSucceed(key string, field string) bool {
	key += ":succeed"
	return hExists(key, field)
}

func makeSucceed(key string, field string, processName string) {
	removeFromProcessing(field)
	queue := key + ":succeed"
	redisClient.HSet(queue, field, getDateTime())
	removeStatusCodeFailed(key, field, 403)
	removeStatusCodeFailed(key, field, 404)
	removeStatusCodeFailed(key, field, 410)
	removeStatusCodeFailed(key, field, 500)
	removeStatusCodeFailed(key, field, 502)
	removeFailed(key, field)
	removeFromQueue(key, field, processName)
}

func getSucceedNum(key string) int64 {
	key += ":succeed"
	return lLen(key)
}

func makeFailed(key string, field string, content string) {
	key += ":failed"
	redisClient.HSet(key, field, content)
}

func removeFailed(key string, field string) {
	key += ":failed"
	redisClient.HDel(key, field)
}

func getFailedNum(key string) int64 {
	key += ":failed"
	return lLen(key)
}

func makeStatusCodeFailed(key string, statusCode int, field string, content string) {
	key += ":" + strconv.Itoa(statusCode)
	redisClient.HSet(key, field, content)
}

func removeStatusCodeFailed(key string, field string, statusCode int) {
	key += ":" + strconv.Itoa(statusCode)
	redisClient.HDel(key, field)
}

func getStatusCodedFailedNum(key string, statusCode int) int64 {
	key += ":" + strconv.Itoa(statusCode)
	return lLen(key)
}

func pushToQueue(key string, content string, processName string) bool {
	if isInQueue(key, content, processName) {
		return false
	}

	queueKey := key + ":queue"
	addIntoQueue(key, content)
	redisClient.LPush(queueKey, content)
	return true
}

func popFromQueue(key string) ([]string, error) {
	timeout := 1 * time.Second
	key += ":queue"
	return redisClient.BRPop(timeout, key).Result()
}

func popFromQueueStatusCode(key string, statusCode int) ([]string, error) {
	timeout := 1 * time.Second
	key += ":queue:" + strconv.Itoa(statusCode)
	return redisClient.BRPop(timeout, key).Result()
}

func getQueueNum(key string) int64 {
	key += ":queue"
	num, err := redisClient.LLen(key).Result()
	if err != nil {
		num = 0
	}
	return num
}

func lLen(key string) int64 {
	num, err := redisClient.HLen(key).Result()
	if err != nil {
		num = 0
	}
	return num
}

func queueExists(key string) int64 {
	key += ":queue"
	num, err := redisClient.Exists(key).Result()
	if err != nil {
		num = 0
	}
	return num
}

func pushToQueueForStatusCodeRetry(key string, statusCode int, content string) {
	key += ":queue:" + strconv.Itoa(statusCode)
	redisClient.LPush(key, content)
}

func isInQueue(key string, hashKey string, processName string) bool {
	key += ":queued"
	if hExists(key, hashKey) {
		fmt.Println(processName, "Queued", key, mirrorUrl(hashKey))
		return true
	}
	return false
}

func removeFromQueue(key string, content string, processName string) {
	key += ":queued"
	fmt.Println(processName, "Queued Remove", key, content)
	redisClient.HDel(key, content)
}

func addIntoQueue(key string, content string) {
	key += ":queued"
	redisClient.HSet(key, content, getDateTime())
	redisClient.Expire(key, 100*time.Second)
}

func addIntoProcessing(path string) {
	redisClient.HSet(processingKey, path, getDateTime())
	redisClient.Expire(processingKey, 70*time.Second)
}

func removeFromProcessing(path string) {
	redisClient.HDel(processingKey, path)
}
