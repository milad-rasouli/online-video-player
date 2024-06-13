package store

import (
	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/model"
	"github.com/redis/rueidis"
)

const (
	AddChatMessageScript = "redis.call('HSET', KEYS[1], 'user_id', ARGV[1], 'message_text', ARGV[2], 'timestamp', ARGV[3]) redis.call('EXPIRE', KEYS[1], 60) redis.call('LPUSH', 'messages', KEYS[1])"
	GetChatMessageScript = " local messages = redis.call('LRANGE', KEYS[1], 0, -1) local valid_messages = {} for i, message_key in ipairs(messages) do if redis.call('EXISTS', message_key) == 1 then local message = redis.call('HGETALL', message_key) table.insert(valid_messages, message) else redis.call('LREM', KEYS[1], 0, message_key) end end return valid_messages "
)

type RedisMessageStore struct {
	cfg    config.Config
	client rueidis.Client
}

func NewRedisMessageStore(cfg config.Config) (*RedisMessageStore, error) {
	redisClient, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{cfg.RedisAddress}})
	if err != nil {
		return &RedisMessageStore{}, err
	}
	defer redisClient.Close()

	//TODO: making the lua script

	return &RedisMessageStore{
		cfg:    cfg,
		client: redisClient,
	}, nil
}

func (m *RedisMessageStore) Save(msg model.Message) error {
	return nil
}

func (m *RedisMessageStore) GetAll() ([]model.Message, error) {
	return []model.Message{}, nil
}
