package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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
	DownloadDistention = "./output.mkv"
)

type Video struct {
	Cfg   config.Config
	Store store.UserAndVideoStore
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
		log.Printf("set video %s", string(c.BodyRaw())) //TODO: remove this line
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
		log.Printf("Set valid video %+v\n", vc)
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
	log.Printf("video PostGetVideo user is %s", fullName)
	usrVideoInto, err := u.Store.GetUserVideoInfo(c.Context(), model.User{FullName: fullName})
	if err != nil {
		log.Printf("PostGetVideoControllers store error")
		return c.SendStatus(fiber.StatusBadRequest)
	}
	return c.JSON(usrVideoInto)
}

type DownloadProgress struct {
	TotalSize    uint64
	ReceivedSize uint64
	StartTime    time.Time
}

func (wc *DownloadProgress) Write(p []byte) (int, error) {
	n := len(p)
	wc.ReceivedSize += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc *DownloadProgress) PrintProgress() {
	duration := time.Since(wc.StartTime)
	speed := float64(wc.ReceivedSize) / duration.Seconds()
	percent := float64(wc.ReceivedSize) / float64(wc.TotalSize) * 100
	timeLeft := time.Duration(float64(wc.TotalSize-wc.ReceivedSize)/speed) * time.Second
	fmt.Printf("\r%s", strings.Repeat(" ", 50))
	fmt.Printf("\rDownloading... %.2f%% complete, speed: %.2f bytes/sec, time left: %v", percent, speed, timeLeft)
}

// "https://dl5.freeserver.top/www2/film/animation/Weekends.2017.480p.DigiMoviez.mkv?md5=Gr9cGCfzCjt753FRU3VrbQ&expires=1719219153"
func (u *Video) download(url model.UploadedVideo) {
	file, err := os.Create(DownloadDistention)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	resp, err := http.Get(url.URL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	progress := &DownloadProgress{
		TotalSize: uint64(resp.ContentLength),
		StartTime: time.Now(),
	}
	if _, err = io.Copy(file, io.TeeReader(resp.Body, progress)); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print("\nDownload complete.\n")
}

func (u *Video) uuidResponse(uuid string) map[string]string {
	return map[string]string{
		"uuid": uuid,
	}
}
func (u *Video) PostUpload(c *fiber.Ctx) error {
	{
		up, err := u.Store.GetUploadedVideo(c.Context())
		if (err != nil && !errors.Is(err, store.UploadedURLIsEmpty)) || (err != nil && !errors.Is(err, store.UploadedUUIDIsEmpty)) {
			log.Printf("PostUpload store error %s\n", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if len(up.UUID) > 0 {
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
			defer u.Store.RemoveUploadedVideo(context.TODO())
			u.download(model.UploadedVideo{
				URL:  rup.URL,
				UUID: rup.UUID,
			})
		}()
	}
	return c.SendStatus(fiber.StatusOK)
}

func (u *Video) PostUploadStatus(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func (u *Video) Video(c *fiber.Ctx) error {
	videoPath := DefaultVideoPath
	fasthttp.ServeFile(c.Context(), videoPath)

	return nil
}

func (u *Video) Register(c fiber.Router) {
	c.Post("/upload", u.PostUpload)
	c.Post("/upload-status", u.PostUploadStatus)
	c.Post("/set", u.PostSetVideoControllers)
	c.Post("/get", u.PostGetVideoControllers)
	c.Get("/", u.Video)
}
