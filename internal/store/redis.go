package store

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/model"
	"github.com/redis/rueidis"
)

const (
	AddChatMessageScript = "redis.call('HSET', KEYS[1], 'user_id', ARGV[1], 'message_text', ARGV[2], 'timestamp', ARGV[3]) redis.call('EXPIRE', KEYS[1], 60) redis.call('LPUSH', 'messages', KEYS[1]) return 0"
	GetChatMessageScript = " local messages = redis.call('LRANGE', KEYS[1], 0, -1) local valid_messages = {} for i, message_key in ipairs(messages) do if redis.call('EXISTS', message_key) == 1 then local message = redis.call('HGETALL', message_key) table.insert(valid_messages, message) else redis.call('LREM', KEYS[1], 0, message_key) end end return valid_messages "
)

type RedisMessageStore struct {
	cfg          config.Config
	client       rueidis.Client
	saveScript   *rueidis.Lua
	getAllScript *rueidis.Lua
	messageID    uint64
	mu           sync.RWMutex
}
type DisposeFunc func()

func NewRedisMessageStore(cfg config.Config) (*RedisMessageStore, DisposeFunc, error) {
	redisClient, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{cfg.RedisAddress}})
	if err != nil {
		return &RedisMessageStore{}, nil, err
	}

	saveScript := rueidis.NewLuaScript(AddChatMessageScript)
	getAllScript := rueidis.NewLuaScript(GetChatMessageScript)

	return &RedisMessageStore{
			cfg:          cfg,
			client:       redisClient,
			saveScript:   saveScript,
			getAllScript: getAllScript,
			messageID:    1,
		}, func() {
			redisClient.Close()
		}, nil
}

func (m *RedisMessageStore) modelMessageToRedis(msg model.Message) []string {
	return []string{msg.Sender, msg.Body, strconv.FormatInt(msg.CreatedAt.Unix(), 10)}
}

func (m *RedisMessageStore) messageIDToRedis() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messageID++
	return []string{strconv.FormatUint(m.messageID, 10)}
}
func (m *RedisMessageStore) Save(ctx context.Context, msg model.Message) error {
	_, err := m.saveScript.Exec(ctx, m.client, m.messageIDToRedis(), m.modelMessageToRedis(msg)).ToInt64()
	return err
}

func (m *RedisMessageStore) GetAll(ctx context.Context) ([]model.Message, error) {
	var (
		messages []model.Message
	)
	list, err := m.getAllScript.Exec(ctx, m.client, []string{}, []string{}).ToArray()
	if err != nil {
		return messages, err
	}
	fmt.Printf("data is : %+v\n", list)

	return messages, nil
}
