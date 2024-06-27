package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/model"
	"github.com/Milad75Rasouli/online-video-player/internal/request"
	"github.com/Milad75Rasouli/online-video-player/internal/store"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

const (
	DefaultVideoPath   = "./internal/static/video/default.mkv"
	DownloadDistention = "./internal/static/video/temp.mkv"
)

type Video struct {
	Cfg        config.Config
	Store      store.UserAndVideoStore
	cancelFunc context.CancelFunc
}

func (u *Video) PostSetVideoControllers(c *fiber.Ctx) error {
	var (
		vc  request.VideoControllers
		err error
	)

	var userFullName = c.Locals("userFullName")
	fullName, ok := userFullName.(string)
	if !ok {
		log.Printf("PostSetVideoControllers invalid userFullName error")
		return c.SendStatus(fiber.StatusBadRequest)
	}
	{
		err = c.BodyParser(&vc)
		if err != nil {
			log.Printf("setVideoControllers body parse error %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		err = vc.Valid()
		if err != nil {
			log.Printf("setVideoControllers body parse validation error %s", err)
			return c.SendStatus(fiber.StatusBadRequest)
		}

		if fullName != vc.User {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		// log.Printf("Set valid video %+v\n", vc)
	}

	{
		err = u.Store.SaveUserVideoInfo(c.Context(), model.User{FullName: vc.User}, model.VideoControllers{
			Pause:    vc.Pause,
			Timeline: vc.Timeline,
		})
		if err != nil {
			log.Printf("setVideoControllers save error %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}
	return nil
}
func (u *Video) PostGetVideoControllers(c *fiber.Ctx) error {
	var userFullName = c.Locals("userFullName")
	fullName, ok := userFullName.(string)
	if !ok {
		log.Printf("PostGetVideoControllers invalid userFullName error")
		return c.SendStatus(fiber.StatusBadRequest)
	}
	// log.Printf("video PostGetVideo user is %s", fullName)
	usrVideoInto, err := u.Store.GetUserVideoInfo(c.Context(), model.User{FullName: fullName})
	if err != nil {
		log.Printf("PostGetVideoControllers store error %s\n", err.Error())
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.JSON(usrVideoInto)
}

type DownloadProgress struct {
	TotalSize    uint64
	ReceivedSize uint64
	StartTime    time.Time
	Store        store.UserAndVideoStore
	User         string
}

func (wc *DownloadProgress) Write(p []byte) (int, error) {
	n := len(p)
	wc.ReceivedSize += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc *DownloadProgress) PrintProgress() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	duration := time.Since(wc.StartTime)
	speed := float64(wc.ReceivedSize) / duration.Seconds()
	percent := float64(wc.ReceivedSize) / float64(wc.TotalSize) * 100
	timeLeft := time.Duration(float64(wc.TotalSize-wc.ReceivedSize)/speed) * time.Second
	// fmt.Printf("\r%s", strings.Repeat(" ", 50))
	// fmt.Printf("\rDownloading... %.2f%% complete, speed: %.2f bytes/sec, time left: %v", percent, speed, timeLeft)
	wc.Store.SaveDownloadVideoStatus(ctx, model.DownloadStatus{
		TotalSize:    wc.TotalSize,
		ReceivedSize: wc.ReceivedSize,
		StartTime:    wc.StartTime.Unix(),
		Percent:      percent,
		Speed:        speed,
		User:         wc.User,
		TimeLeft:     timeLeft.String(),
	})
}

// https://dl5.freeserver.top/www2/film/animation/Weekends.2017.480p.DigiMoviez.mkv?md5=Gr9cGCfzCjt753FRU3VrbQ&expires=1719219153
func (u *Video) download(ctx context.Context, url model.UploadedVideo, user string) {
	file, err := os.Create(DownloadDistention)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.URL, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	progress := &DownloadProgress{
		TotalSize: uint64(resp.ContentLength),
		StartTime: time.Now(),
		Store:     u.Store,
		User:      user,
	}
	if _, err = io.Copy(file, io.TeeReader(resp.Body, progress)); err != nil {
		fmt.Println(err)
		return
	}

	// fmt.Print("\nDownload complete.\n")
	err = os.Rename(DownloadDistention, DefaultVideoPath)
	if err != nil {
		fmt.Println("Error moving file:", err)
		return
	}

	// fmt.Print("\nFile moved successfully.\n")
}

func (u *Video) uuidResponse(uuid string) map[string]string {
	return map[string]string{
		"uuid": uuid,
	}
}
func (u *Video) PostUpload(c *fiber.Ctx) error {
	var userFullName = c.Locals("userFullName")
	fullName, ok := userFullName.(string)
	if !ok {
		log.Printf("PostUpload invalid userFullName error")
		return c.SendStatus(fiber.StatusBadRequest)
	}
	// log.Println("upload Video is called")
	{
		up, _ := u.Store.GetUploadedVideo(c.Context())
		// if (err != nil && !errors.Is(err, store.UploadedURLIsEmpty)) || (err != nil && !errors.Is(err, store.UploadedUUIDIsEmpty)) {
		// 	log.Printf("PostUpload store error %s\n", err)
		// 	return c.SendStatus(fiber.StatusInternalServerError)
		// } //TODO: Handel this error part
		if len(up.UUID) > 0 {
			log.Println("UploadVideo returned uuid ", u.uuidResponse(up.UUID))
			return c.JSON(u.uuidResponse(up.UUID))
		}
	}
	var rup request.UploadedVideo
	{
		c.BodyParser(&rup)
		err := rup.Valid()
		if err != nil {
			log.Printf("PostUpload json parse error %s", err)
			return c.SendStatus(fiber.StatusBadRequest)
		}

	}
	{
		u.Store.SaveUploadedVideo(c.Context(), model.UploadedVideo{
			URL:  rup.URL,
			UUID: rup.UUID,
		})
		go func() {
			var downloadCtx context.Context
			downloadCtx, u.cancelFunc = context.WithCancel(context.Background())
			defer u.Store.RemoveUploadedVideo(context.Background())
			defer u.Store.RemoveDownloadVideoStatus(context.Background())
			u.download(downloadCtx, model.UploadedVideo{
				URL:  rup.URL,
				UUID: rup.UUID,
			}, fullName)
		}()
	}
	return c.SendStatus(fiber.StatusOK)
}

func (u *Video) PostUploadStatus(c *fiber.Ctx) error {
	ds, _ := u.Store.GetDownloadVideoStatus(c.Context())
	if ds.User == "" {
		return c.SendStatus(fiber.StatusOK)
	}
	// if err != nil {
	// 	log.Printf("PostUploadStatus store error") //TODO: Fix the error handling
	// 	return c.SendStatus(fiber.StatusInternalServerError)
	// }

	return c.JSON(ds)
}

func (u *Video) PostCancelUpload(c *fiber.Ctx) error {
	if u.cancelFunc != nil {
		u.cancelFunc()
	} else {
		log.Println("early download cancel function")
		return c.SendStatus(fiber.StatusTooEarly)
	}
	u.Store.RemoveUploadedVideo(c.Context())
	u.Store.RemoveDownloadVideoStatus(c.Context())
	return c.SendStatus(fiber.StatusOK)
}
func (u *Video) Video(c *fiber.Ctx) error {
	videoPath := DefaultVideoPath
	fasthttp.ServeFile(c.Context(), videoPath)

	return nil
}

func (u *Video) Register(c fiber.Router) {
	c.Post("/upload", u.PostUpload)
	c.Post("/cancel-upload", u.PostCancelUpload)
	c.Post("/upload-status", u.PostUploadStatus)
	c.Post("/set", u.PostSetVideoControllers)
	c.Post("/get", u.PostGetVideoControllers)
	c.Get("/", u.Video)
}
