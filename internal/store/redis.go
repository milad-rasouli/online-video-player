package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/model"
	"github.com/google/uuid"
	"github.com/redis/rueidis"
)

const (
	AddChatMessageScript     = "redis.call('HSET', KEYS[1], 'user_id', ARGV[1], 'message_text', ARGV[2], 'timestamp', ARGV[3]) redis.call('EXPIRE', KEYS[1], %s) redis.call('RPUSH', 'messages', KEYS[1]) return 0"
	GetChatMessageScript     = " local messages = redis.call('LRANGE', KEYS[1], 0, -1) local valid_messages = {} for i, message_key in ipairs(messages) do if redis.call('EXISTS', message_key) == 1 then local message = redis.call('HGETALL', message_key) table.insert(valid_messages, message) else redis.call('LREM', KEYS[1], 0, message_key) end end return valid_messages "
	MAX_LENGTH_OF_MESSAGE    = 6
	ExpireUserVideoInfoAfter = 120
	ExpireUploadedURLTime    = 1200
	ExpireDownloadStatusTime = 15
)

var (
	TimelineIsEmpty                   = errors.New("Timeline is empty")
	PauseIsEmpty                      = errors.New("Pause is empty")
	UploadedURLIsEmpty                = errors.New("Uploaded url is empty")
	UploadedUUIDIsEmpty               = errors.New("Uploaded url is empty")
	DownloadStatusParseTotalSizeError = errors.New("Download status total size parse error")
	DownloadStatusTotalSizeIsEmpty    = errors.New("Download status total size is empty")
	DownloadStatusStartTimeIsEmpty    = errors.New("Download status start time is empty")
	DownloadStatusParseStartTimeError = errors.New("Download status start time parse error")
	DownloadStatusReceivedSizeError   = errors.New("Download status received size parse error")
	DownloadReceivedSizeIsEmpty       = errors.New("Download status received size is empty")
	DownloadStatusSpeedError          = errors.New("Download status speed parse error")
	DownloadSpeedEmpty                = errors.New("Download status speed is empty")
	DownloadStatusPercentError        = errors.New("Download status percent parse error")
	DownloadPercentEmpty              = errors.New("Download status percent is empty")
	DownloadTimeLeftEmpty             = errors.New("Download status time left is empty")
	DownloadStatusUserIsEmpty         = errors.New("Download status user is empty")
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

type RedisUserAndVideStore struct {
	cfg    config.Config
	client rueidis.Client
}

func NewRedisUserAndVideStore(cfg config.Config) (*RedisUserAndVideStore, DisposeFunc, error) {
	redisClient, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{cfg.RedisAddress}})
	if err != nil {
		return &RedisUserAndVideStore{}, nil, err
	}

	return &RedisUserAndVideStore{
			cfg:    cfg,
			client: redisClient,
		}, func() {
			redisClient.Close()
		}, nil
}

func (r *RedisUserAndVideStore) userVideoInfoID(user model.User) string {
	return "user:" + user.FullName
}

func (r *RedisUserAndVideStore) SaveUserVideoInfo(ctx context.Context, user model.User, vc model.VideoControllers) error {
	var (
		err error
		key = r.userVideoInfoID(user)
	)

	err = r.client.Do(ctx, r.client.B().
		Hset().
		Key(key).
		FieldValue().
		FieldValue("timeline", vc.Timeline).
		FieldValue("pause", strconv.FormatBool(vc.Pause)).
		Build()).Error()
	if err != nil {
		return err
	}
	err = r.client.Do(ctx, r.client.B().Expire().Key(key).Seconds(ExpireUserVideoInfoAfter).Build()).Error()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisUserAndVideStore) GetUserVideoInfo(ctx context.Context, user model.User) (model.VideoControllers, error) {
	var (
		key = r.userVideoInfoID(user)
		vc  model.VideoControllers
	)

	data, err := r.client.Do(ctx, r.client.B().Hgetall().Key(key).Build()).AsStrMap()
	if err != nil {
		return model.VideoControllers{}, err
	}
	timeline, ok := data["timeline"]
	if ok {
		vc.Timeline = timeline
	} else {
		return model.VideoControllers{}, TimelineIsEmpty
	}
	pause, ok := data["pause"]
	if ok {
		p, err := strconv.ParseBool(pause)
		if err != nil {
			return model.VideoControllers{}, err
		}
		vc.Pause = p
	} else {
		return model.VideoControllers{}, PauseIsEmpty
	}
	return vc, err
}
func (r *RedisUserAndVideStore) uploadedVideoID() string {
	return "uploadedURL"
}
func (r *RedisUserAndVideStore) SaveUploadedVideo(ctx context.Context, url model.UploadedVideo) error {
	cmds := make(rueidis.Commands, 0, 2)
	cmds = append(cmds, r.client.B().Hset().Key(r.uploadedVideoID()).FieldValue().FieldValue("url", url.URL).FieldValue("uuid", url.UUID).Build())
	cmds = append(cmds, r.client.B().Expire().Key(r.uploadedVideoID()).Seconds(ExpireUploadedURLTime).Build())

	for _, resp := range r.client.DoMulti(ctx, cmds...) {
		if err := resp.Error(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisUserAndVideStore) GetUploadedVideo(ctx context.Context) (model.UploadedVideo, error) {
	var url model.UploadedVideo
	data, err := r.client.Do(ctx, r.client.B().Hgetall().Key(r.uploadedVideoID()).Build()).AsStrMap()
	if err != nil {
		return model.UploadedVideo{}, err
	}

	u, ok := data["url"]
	if ok {
		url.URL = u
	} else {
		return model.UploadedVideo{}, UploadedURLIsEmpty
	}

	uu, ok := data["uuid"]
	if ok {
		url.UUID = uu
	} else {
		return model.UploadedVideo{}, UploadedUUIDIsEmpty
	}

	return url, nil
}
func (r *RedisUserAndVideStore) RemoveUploadedVideo(ctx context.Context) error {
	return r.client.Do(ctx, r.client.B().Del().Key(r.uploadedVideoID()).Build()).Error()
}
func (r *RedisUserAndVideStore) downloadStatusID() string {
	return "downloadStatus"
}
func (r *RedisUserAndVideStore) SaveDownloadVideoStatus(ctx context.Context, status model.DownloadStatus) error {
	cmds := make(rueidis.Commands, 0, 2)
	cmds = append(cmds, r.client.B().
		Hset().
		Key(r.downloadStatusID()).
		FieldValue().
		FieldValue("user", status.User).
		FieldValue("totalSize", strconv.FormatUint(status.TotalSize, 10)).
		FieldValue("startTime", strconv.FormatInt(status.StartTime, 10)).
		FieldValue("receivedSize", strconv.FormatUint(status.ReceivedSize, 10)).
		FieldValue("speed", strconv.FormatFloat(status.Speed, 'E', -1, 32)).
		FieldValue("percent", strconv.FormatFloat(status.Percent, 'E', -1, 32)).
		FieldValue("timeLeft", status.TimeLeft).
		Build())
	cmds = append(cmds, r.client.B().Expire().Key(r.downloadStatusID()).Seconds(ExpireDownloadStatusTime).Build())

	for _, resp := range r.client.DoMulti(ctx, cmds...) {
		if err := resp.Error(); err != nil {
			return err
		}
	}
	return nil
}
func (r *RedisUserAndVideStore) GetDownloadVideoStatus(ctx context.Context) (model.DownloadStatus, error) {
	var status model.DownloadStatus
	data, err := r.client.Do(ctx, r.client.B().Hgetall().Key(r.downloadStatusID()).Build()).AsStrMap()
	if err != nil {
		return model.DownloadStatus{}, err
	}

	usr, ok := data["user"]
	if ok {
		status.User = usr
	} else {
		return model.DownloadStatus{}, DownloadStatusUserIsEmpty
	}
	t, ok := data["totalSize"]
	if ok {
		status.TotalSize, err = strconv.ParseUint(t, 10, 64)
		if err != nil {
			return model.DownloadStatus{}, DownloadStatusParseTotalSizeError
		}
	} else {
		return model.DownloadStatus{}, DownloadStatusTotalSizeIsEmpty
	}

	s, ok := data["startTime"]
	if ok {
		status.StartTime, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return model.DownloadStatus{}, DownloadStatusParseStartTimeError
		}
	} else {
		return model.DownloadStatus{}, DownloadStatusStartTimeIsEmpty
	}

	rs, ok := data["receivedSize"]
	if ok {
		status.ReceivedSize, err = strconv.ParseUint(rs, 10, 64)
		if err != nil {
			return model.DownloadStatus{}, DownloadStatusReceivedSizeError
		}
	} else {
		return model.DownloadStatus{}, DownloadReceivedSizeIsEmpty
	}

	sd, ok := data["speed"]
	if ok {
		status.Speed, err = strconv.ParseFloat(sd, 64)
		if err != nil {
			return model.DownloadStatus{}, DownloadStatusSpeedError
		}
	} else {
		return model.DownloadStatus{}, DownloadSpeedEmpty
	}

	pt, ok := data["percent"]
	if ok {
		status.Percent, err = strconv.ParseFloat(pt, 64)
		if err != nil {
			return model.DownloadStatus{}, DownloadStatusPercentError
		}
	} else {
		return model.DownloadStatus{}, DownloadPercentEmpty
	}

	tl, ok := data["timeLeft"]
	if ok {
		status.TimeLeft = tl
	} else {
		return model.DownloadStatus{}, DownloadTimeLeftEmpty
	}

	return status, nil
}
func (r *RedisUserAndVideStore) RemoveDownloadVideoStatus(ctx context.Context) error {
	return r.client.Do(ctx, r.client.B().Del().Key(r.downloadStatusID()).Build()).Error()
}

// func (r *RedisUserAndVideStore) SaveUser(ctx context.Context, user model.User) error {
// 	return r.client.Do(ctx, r.client.B().Rpush().Key(r.userListID()).Element(user.FullName).Build()).Error()
// }
// func (r *RedisUserAndVideStore) RemoveAllUser(ctx context.Context) error {
// 	return r.client.Do(ctx, r.client.B().Del().Key(r.userListID()).Build()).Error()
// }
// func (r *RedisUserAndVideStore) GetAllUser(ctx context.Context) ([]model.User, error) {
// 	var users []model.User
// 	data, err := r.client.Do(ctx, r.client.B().Lrange().Key(r.userListID()).Start(0).Stop(-1).Build()).AsStrSlice()
// 	if err != nil {
// 		return []model.User{}, err
// 	}
// 	for _, i := range data {
// 		users = append(users, model.User{FullName: i})
// 	}
// 	return users, nil
// }
// func (r *RedisUserAndVideStore) currentVideoID(user model.User) string {
// 	return "currentVideoController:" + user.FullName
// }
// func (r *RedisUserAndVideStore) SaveCurrentVideo(ctx context.Context, vc model.VideoControllers) error {
// 	var (
// 		users []model.User
// 		err   error
// 	)
// 	{
// 		users, err = r.GetAllUser(ctx)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	{
// 		usersSize := len(users)
// 		cmds := make(rueidis.Commands, 0, usersSize)
// 		for _, user := range users {
// 			key := r.currentVideoID(user)
// 			cmds = append(cmds, r.client.
// 				B().
// 				Hset().
// 				Key(key).
// 				FieldValue().
// 				FieldValue("pause", strconv.FormatBool(vc.Pause)).
// 				FieldValue("timeline", vc.Timeline).
// 				FieldValue("movie", vc.Movie).
// 				Build())
// 			cmds = append(cmds, r.client.
// 				B().
// 				Expire().
// 				Key(key).
// 				Seconds(5).
// 				Build())
// 		}
// 		for _, resp := range r.client.DoMulti(ctx, cmds...) {
// 			if err := resp.Error(); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }
// func (r *RedisUserAndVideStore) GetCurrentVideo(ctx context.Context, user model.User) (model.VideoControllers, error) {
// 	data, err := r.client.Do(ctx, r.client.B().Hgetall().Key(r.currentVideoID(user)).Build()).AsStrMap()
// 	if err != nil {
// 		return model.VideoControllers{}, err
// 	}

// 	pause, err := strconv.ParseBool(data["pause"])
// 	if err != nil {
// 		return model.VideoControllers{}, err
// 	}
// 	return model.VideoControllers{
// 		Timeline: data["timeline"],
// 		Movie:    data["movie"],
// 		Pause:    pause,
// 	}, nil
// }
// func (r *RedisUserAndVideStore) RemoveCurrentVideo(ctx context.Context, user model.User) error {
// 	return r.client.Do(ctx, r.client.B().Del().Key(r.currentVideoID(user)).Build()).Error()
// }

// func (r *RedisUserAndVideStore) playlistID() string {
// 	return "playlist"
// }
// func (r *RedisUserAndVideStore) SaveToPlaylist(ctx context.Context, p model.Playlist) error {
// 	return r.client.Do(ctx, r.client.B().Rpush().Key(r.playlistID()).Element(p.Item).Build()).Error()
// }
// func (r *RedisUserAndVideStore) GetPlaylist(ctx context.Context) ([]model.Playlist, error) {
// 	data, err := r.client.Do(ctx, r.client.B().Lrange().Key(r.playlistID()).Start(0).Stop(-1).Build()).AsStrSlice()
// 	if err != nil {
// 		return []model.Playlist{}, err
// 	}

// 	fmt.Printf("%+v\n", data)
// 	playlist := []model.Playlist{}
// 	for _, item := range data {
// 		playlist = append(playlist, model.Playlist{Item: item})
// 	}
// 	return playlist, nil
// }

// func (r *RedisUserAndVideStore) RemovePlaylist(ctx context.Context) error {
// 	return r.client.Do(ctx, r.client.B().Del().Key(r.playlistID()).Build()).Error()
// }
