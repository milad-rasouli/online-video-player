package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestRedis(t *testing.T) {
	cfg := config.Config{
		RedisAddress: "127.0.0.1:6379",
		RedisChatExp: "10",
	}
	redis, disposeRedis, err := NewRedisMessageStore(cfg)
	assert.NoError(t, err)
	defer disposeRedis()

	messages := [...]model.Message{
		{
			Sender:    "foo",
			Body:      "hahahaha",
			CreatedAt: time.Now(),
		},
		{
			Sender:    "boo",
			Body:      "HEHEHE",
			CreatedAt: time.Now().Add(time.Hour),
		},
		{
			Sender:    "bazz",
			Body:      "HAHAHAHA",
			CreatedAt: time.Now().Add(time.Minute),
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	{
		for _, msg := range messages {
			redis.Save(ctx, msg)
		}
	}

	{
		data, err := redis.GetAll(ctx)
		assert.NoError(t, err)
		// assert.Equal(t, messages, data)
		fmt.Println(data)

	}
}