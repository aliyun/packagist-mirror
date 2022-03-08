package util

import (
	"strconv"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-redis/redis"
)

type Context struct {
	ossBucket *oss.Bucket
	redis     *redis.Client
	packagist *Packagist
	cdn       *CDN
	github    *Github
	mirror    *Mirror
}

func NewContext(conf *Config) (ctx *Context, err error) {
	ossclient, err := oss.New(conf.OSSEndpoint, conf.OSSAccessKeyID, conf.OSSAccessKeySecret)
	if err != nil {
		return
	}

	bucketClient, err := ossclient.Bucket(conf.OSSBucket)
	if err != nil {
		return
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     conf.RedisAddr,
		Password: conf.RedisPassword,
		DB:       conf.RedisDB,
	})

	ctx = &Context{
		ossBucket: bucketClient,
		redis:     redisClient,
		packagist: NewPackagist(conf.RepoUrl, conf.ApiUrl, conf.UserAgent),
		cdn:       NewCDN(conf.BuildCache == "true", conf.MirrorUrl),
		github:    NewGithub(conf.GithubToken, conf.UserAgent),
		mirror:    NewMirror(conf.ProviderUrl, conf.DistUrl, conf.ApiIterationInterval),
	}
	return
}

func (ctx *Context) redisHLen(key string) int64 {
	num, err := ctx.redis.HLen(key).Result()
	if err != nil {
		num = 0
	}
	return num
}

func (ctx *Context) redisSCard(key string) int64 {
	num, err := ctx.redis.SCard(key).Result()
	if err != nil {
		num = 0
	}
	return num
}

func (ctx *Context) countFailed(key string) int64 {
	key += "-failed"
	return ctx.redisHLen(key)
}

func (ctx *Context) countStatusCodedFailed(key string, statusCode int) int64 {
	key += "-" + strconv.Itoa(statusCode)
	return ctx.redisSCard(key)
}
