package store

import (
	"context"
	"encoding/json"
	"fmt"
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

func (r *RedisVideoControllerStore) userListID() string {
	return "userList"
}
func (r *RedisVideoControllerStore) SaveUser(ctx context.Context, user model.User) error {
	return r.client.Do(ctx, r.client.B().Rpush().Key(r.userListID()).Element(user.FullName).Build()).Error()
}
func (r *RedisVideoControllerStore) RemoveAllUser(ctx context.Context) error {
	return r.client.Do(ctx, r.client.B().Del().Key(r.userListID()).Build()).Error()
}
func (r *RedisVideoControllerStore) GetAllUser(ctx context.Context) ([]model.User, error) {
	var users []model.User
	data, err := r.client.Do(ctx, r.client.B().Lrange().Key(r.userListID()).Start(0).Stop(-1).Build()).AsStrSlice()
	if err != nil {
		return []model.User{}, err
	}
	for _, i := range data {
		users = append(users, model.User{FullName: i})
	}
	return users, nil
}

func (r *RedisVideoControllerStore) currentVideoID(user model.User) string {
	return "currentVideoController:" + user.FullName
}
func (r *RedisVideoControllerStore) SaveCurrentVideo(ctx context.Context, vc model.VideoControllers) error {
	var (
		users []model.User
		err   error
	)
	{
		users, err = r.GetAllUser(ctx)
		if err != nil {
			return err
		}
	}
	{
		usersSize := len(users)
		cmds := make(rueidis.Commands, 0, usersSize)
		for _, user := range users {
			key := r.currentVideoID(user)
			cmds = append(cmds, r.client.
				B().
				Hset().
				Key(key).
				FieldValue().
				FieldValue("pause", strconv.FormatBool(vc.Pause)).
				FieldValue("timeline", vc.Timeline).
				FieldValue("movie", vc.Movie).
				Build())
			cmds = append(cmds, r.client.
				B().
				Expire().
				Key(key).
				Seconds(5).
				Build())
		}
		for _, resp := range r.client.DoMulti(ctx, cmds...) {
			if err := resp.Error(); err != nil {
				return err
			}
		}
	}
	return nil
}
func (r *RedisVideoControllerStore) GetCurrentVideo(ctx context.Context, user model.User) (model.VideoControllers, error) {
	data, err := r.client.Do(ctx, r.client.B().Hgetall().Key(r.currentVideoID(user)).Build()).AsStrMap()
	if err != nil {
		return model.VideoControllers{}, err
	}

	pause, err := strconv.ParseBool(data["pause"])
	if err != nil {
		return model.VideoControllers{}, err
	}
	return model.VideoControllers{
		Timeline: data["timeline"],
		Movie:    data["movie"],
		Pause:    pause,
	}, nil
}
func (r *RedisVideoControllerStore) RemoveCurrentVideo(ctx context.Context, user model.User) error {
	return r.client.Do(ctx, r.client.B().Del().Key(r.currentVideoID(user)).Build()).Error()
}

func (r *RedisVideoControllerStore) playlistID() string {
	return "playlist"
}
func (r *RedisVideoControllerStore) SaveToPlaylist(ctx context.Context, p model.Playlist) error {
	return r.client.Do(ctx, r.client.B().Rpush().Key(r.playlistID()).Element(p.Item).Build()).Error()
}
func (r *RedisVideoControllerStore) GetPlaylist(ctx context.Context) ([]model.Playlist, error) {
	data, err := r.client.Do(ctx, r.client.B().Lrange().Key(r.playlistID()).Start(0).Stop(-1).Build()).AsStrSlice()
	if err != nil {
		return []model.Playlist{}, err
	}

	fmt.Printf("%+v\n", data)
	playlist := []model.Playlist{}
	for _, item := range data {
		playlist = append(playlist, model.Playlist{Item: item})
	}
	return playlist, nil
}

func (r *RedisVideoControllerStore) RemovePlaylist(ctx context.Context) error {
	return r.client.Do(ctx, r.client.B().Del().Key(r.playlistID()).Build()).Error()
}
