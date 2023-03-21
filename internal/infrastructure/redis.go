package infrastructure

import "github.com/go-redis/redis/v8"

func NewRedisClient(configName string) *redis.Client {
	config := NewConfig(configName)
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisHost,
		Password: config.RedisPassword,
	})
	return client
}
