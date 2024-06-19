package store

import (
	"context"
	"fmt"
	"log"
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

func TestVideoController(t *testing.T) {
	cfg := config.Config{
		RedisAddress: "127.0.0.1:6379",
	}
	redis, disposeRedis, err := NewRedisVideoControllerStore(cfg)
	assert.NoError(t, err)
	defer disposeRedis()

	m := model.VideoControllers{
		Pause:    false,
		Timeline: "12:12",
		Movie:    "foo bar baz",
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	{
		err := redis.SaveCurrentVideo(ctx, m)
		assert.NoError(t, err)
	}

	{
		data, err := redis.GetCurrentVideo(ctx)
		assert.NoError(t, err)
		assert.Equal(t, m, data)
		log.Printf("%+v\n", data)
	}

	p := []model.Playlist{
		{Item: "foo foo"},
		{Item: "bar bar"},
		{Item: "baz baz"},
	}
	{
		err := redis.RemovePlaylist(ctx)
		assert.NoError(t, err)
	}
	{
		for _, i := range p {
			err := redis.SaveToPlaylist(ctx, i)
			assert.NoError(t, err)
		}
	}

	{
		fp, err := redis.GetPlaylist(ctx)
		assert.NoError(t, err)
		assert.Equal(t, p, fp)
		log.Printf("test: %+v\n", fp)
	}
}

// func TestVideoControllerEmpty(t *testing.T) {
// 	cfg := config.Config{ //TODO: run the redis database with dockertest
// 		RedisAddress: "127.0.0.1:6379",
// 	}
// 	redis, disposeRedis, err := NewRedisVideoControllerStore(cfg)
// 	assert.NoError(t, err)
// 	defer disposeRedis()

// 	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
// 	defer cancel()
// 	{
// 		data, err := redis.GetCurrentVideo(ctx)
// 		assert.NoError(t, err)
// 		fmt.Printf("Video %+v\n", data)
// 	}
// 	{
// 		err := redis.RemovePlaylist(ctx)
// 		assert.NoError(t, err)
// 	}
// 	{
// 		data, err := redis.GetPlaylist(ctx)
// 		assert.NoError(t, err)
// 		fmt.Printf("Playlist %+v\n", data)
// 	}
// }
