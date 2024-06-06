package fwRedis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"go-qirania/config"
	"go-qirania/utils/milog"
	"sync"
)

type rdsClient struct {
	client redis.UniversalClient
}

var onceRedis sync.Once
var rdsQueue *rdsClient

func RedisInit() {
	if rdsQueue != nil {
		return
	}

	onceRedis.Do(func() {
		rdsQueue = new(rdsClient)
		rdsQueue.connectDB(config.Conf.Redis)
	})
}

func RedisQueue() redis.UniversalClient {
	return rdsQueue.client
}

func (r *rdsClient) connectDB(conf config.RedisConfig) {
	ctx := context.Background()

	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    conf.Host,
		Password: conf.Password,
		DB:       conf.DB,
	})

	pong, err := client.Ping(ctx).Result()
	if err != nil {
		milog.Fatal("redis connect ping failed, err:", err)
	} else {
		milog.Debug("redis connect ping response:", "pong", pong)
		milog.Info("Redis DB: ", conf.DB)
		r.client = client
	}
}
