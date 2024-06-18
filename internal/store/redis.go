package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/model"
	"github.com/google/uuid"
	"github.com/redis/rueidis"
)

const (
	AddChatMessageScript  = "redis.call('HSET', KEYS[1], 'user_id', ARGV[1], 'message_text', ARGV[2], 'timestamp', ARGV[3]) redis.call('EXPIRE', KEYS[1], %s) redis.call('RPUSH', 'messages', KEYS[1]) return 0"
	GetChatMessageScript  = " local messages = redis.call('LRANGE', KEYS[1], 0, -1) local valid_messages = {} for i, message_key in ipairs(messages) do if redis.call('EXISTS', message_key) == 1 then local message = redis.call('HGETALL', message_key) table.insert(valid_messages, message) else redis.call('LREM', KEYS[1], 0, message_key) end end return valid_messages "
	MAX_LENGTH_OF_MESSAGE = 6
)

type RedisMessageStore struct {
	cfg          config.Config
	client       rueidis.Client
	saveScript   *rueidis.Lua
	getAllScript *rueidis.Lua
}
type DisposeFunc func()

func NewRedisMessageStore(cfg config.Config) (*RedisMessageStore, DisposeFunc, error) {
	redisClient, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{cfg.RedisAddress}})
	if err != nil {
		return &RedisMessageStore{}, nil, err
	}

	timeModifiedAddChatMessageScript := fmt.Sprintf(AddChatMessageScript, cfg.RedisChatExp)
	fmt.Println("AddChatMessageScript: ", timeModifiedAddChatMessageScript)
	saveScript := rueidis.NewLuaScript(timeModifiedAddChatMessageScript)
	getAllScript := rueidis.NewLuaScript(GetChatMessageScript)

	return &RedisMessageStore{
			cfg:          cfg,
			client:       redisClient,
			saveScript:   saveScript,
			getAllScript: getAllScript,
		}, func() {
			redisClient.Close()
		}, nil
}

func (m *RedisMessageStore) modelMessageToRedis(msg model.Message) []string {
	return []string{msg.Sender, msg.Body, strconv.FormatInt(msg.CreatedAt.Unix(), 10)}
}

func (m *RedisMessageStore) messageIDToRedis() []string {
	return []string{"msg" + uuid.New().String()}
}
func (m *RedisMessageStore) Save(ctx context.Context, msg model.Message) error {
	_, err := m.saveScript.Exec(ctx, m.client, m.messageIDToRedis(), m.modelMessageToRedis(msg)).ToInt64()
	return err
}

func (m *RedisMessageStore) messageListID() []string {
	return []string{"messages"}
}

func (m *RedisMessageStore) GetAll(ctx context.Context) ([]model.Message, error) {
	var messages []model.Message
	list, err := m.getAllScript.Exec(ctx, m.client, m.messageListID(), []string{}).ToArray()
	if err != nil {
		return messages, err
	}
	for _, msg := range list {
		data, err := msg.ToArray()
		if err != nil {
			return messages, err
		}
		if len(data) > MAX_LENGTH_OF_MESSAGE {
			continue
		}

		sender, err := parseJSON(data[1].String())
		if err != nil {
			return messages, err
		}

		body, err := parseJSON(data[3].String())
		if err != nil {
			return messages, err
		}

		ct, err := parseJSON(data[5].String())
		if err != nil {
			return messages, err
		}

		createAt, err := strconv.ParseInt(ct.Value, 10, 64)
		if err != nil {
			return messages, err
		}

		message := model.Message{
			Sender:    sender.Value,
			Body:      body.Value,
			CreatedAt: time.Unix(createAt, 0),
		}
		messages = append(messages, message)
	}
	return messages, nil
}

type ValueType struct {
	Value string `json:"Value"`
	Type  string `json:"Type"`
}

func parseJSON(jsonStr string) (ValueType, error) {
	var blobString ValueType
	err := json.Unmarshal([]byte(jsonStr), &blobString)
	return blobString, err
}

type RedisVideoControllerStore struct {
	cfg    config.Config
	client rueidis.Client
}

func NewRedisVideoControllerStore(cfg config.Config) (*RedisVideoControllerStore, DisposeFunc, error) {
	redisClient, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{cfg.RedisAddress}})
	if err != nil {
		return &RedisVideoControllerStore{}, nil, err
	}

	return &RedisVideoControllerStore{
			cfg:    cfg,
			client: redisClient,
		}, func() {
			redisClient.Close()
		}, nil
}

func (r *RedisVideoControllerStore) CurrentVideoID() string {
	return "CurrentVideoController"
}
func (r *RedisVideoControllerStore) SaveCurrentVideo(ctx context.Context, vc model.VideoControllers) error {
	return r.client.Do(ctx, r.client.
		B().
		Hset().
		Key(r.CurrentVideoID()).
		FieldValue().
		FieldValue("pause", strconv.FormatBool(vc.Pause)).
		FieldValue("timeline", vc.Timeline).
		FieldValue("movie", vc.Movie).
		Build()).Error()

}
func (r *RedisVideoControllerStore) GetCurrentVideo(ctx context.Context, vc model.VideoControllers) error {
	data, err := r.client.Do(ctx, r.client.B().Hgetall().Key(r.CurrentVideoID()).Build()).AsStrSlice()
	if err != nil {
		return err
	}
	log.Println(data)
	// r.client.Do(ctx, r.client.B().Hgetall().Key(r.CurrentVideoID()).Build()).AsStrMap()
	return nil
}

/*
type VideoControllersStore interface {
	SaveCurrentVideo(context.Context, model.VideoControllers) error
	GetCurrentVideo(context.Context) (model.VideoControllers, error)
	SaveToPlaylist(context.Context, model.Playlist) error
	GetPlaylist(context.Context) ([]model.Playlist, error)
	RemovePlaylist(context.Context) error
}
*/
