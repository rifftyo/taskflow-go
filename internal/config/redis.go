package config

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis() (*redis.Client, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	db := 0
	if value:= os.Getenv("REDIS_DB"); value != "" {
		parsedDB, err := strconv.Atoi(value)
		if err != nil || parsedDB < 0 {
			return nil, errors.New("invalid REDIS_DB")
		}
		db = parsedDB
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
		Password: password,
		DB: db,
		DialTimeout: 5 * time.Second,
		ReadTimeout: 3 * time.Second,
		WriteTimeout: 3 *time.Second,
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}