package handler

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/model"
	"github.com/Milad75Rasouli/online-video-player/internal/request"
	"github.com/Milad75Rasouli/online-video-player/internal/store"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

const (
	RedisTimeout = 100
)

// Add more data to this type if needed
type client struct {
	isClosing bool
	mu        sync.Mutex
	unread    bool
}
type Chat struct {
	clients    map[*websocket.Conn]*client
	register   chan *websocket.Conn
	broadcast  chan string
	unregister chan *websocket.Conn
	cfg        config.Config
	redis      store.MessageStore
}

func NewChat(cfg config.Config, redis store.MessageStore) *Chat {
	var clients = make(map[*websocket.Conn]*client)
	var register = make(chan *websocket.Conn)
	var broadcast = make(chan string)
	var unregister = make(chan *websocket.Conn)

	return &Chat{
		cfg:        cfg,
		clients:    clients,
		register:   register,
		broadcast:  broadcast,
		unregister: unregister,
		redis:      redis,
	}
}

func (ch *Chat) GetWebsocket(c *websocket.Conn) {
	defer func() {
		ch.unregister <- c
		c.Close()
	}()
	ch.register <- c
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}
			return
		}

		if messageType == websocket.TextMessage {
			var (
				parsedMsg request.Message
				modelMsg  = model.Message{
					CreatedAt: time.Now(),
				}
				err error
				msg string
			)
			err = json.Unmarshal(message, &parsedMsg)
			if err != nil {
				log.Println("websocket json unmarshal error ", string(message))
				msg, err = ch.ErrorMessage("SYSTEM: Send a valid message please!")
				ch.broadcast <- msg
				if err != nil {
					log.Println("websocket json marshal error message", string(message))
				}
				continue
			}
			err = parsedMsg.Validate()
			log.Printf("GetWebsocket got message from user %+v\n", parsedMsg)
			if err != nil {
				log.Println("websocket invalid request")
				msg, err = ch.ErrorMessage("SYSTEM: " + err.Error())
				ch.broadcast <- msg
				if err != nil {
					log.Println("websocket json marshal error message", string(err.Error()))
				}
				continue
			}

			modelMsg.Body = parsedMsg.Body
			modelMsg.Sender = parsedMsg.Sender
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*RedisTimeout)
			defer cancel()
			ch.redis.Save(ctx, modelMsg)
			finalMsg, err := json.Marshal(modelMsg)
			if err != nil {
				log.Println("websocket unable to make model user")
				msg, err = ch.ErrorMessage("SYSTEM: " + err.Error())
				ch.broadcast <- msg
				if err != nil {
					log.Println("websocket json marshal error message", string(err.Error()))
				}
				continue
			}
			ch.broadcast <- string(finalMsg)
		} else {
			log.Println("websocket message received of type", messageType)
		}
	}
}

func (ch *Chat) ErrorMessage(msg string) (string, error) {
	err := struct {
		Error string `json:"error"`
	}{
		Error: msg,
	}
	errMsg, er := json.Marshal(err)
	return string(errMsg), er
}
func (ch *Chat) ChatWebsocketAcceptorMiddleware(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return c.SendStatus(fiber.StatusUpgradeRequired)
}

// TODO: this should be shutdown properly. in the case i mean
func (ch *Chat) runHub() {
	for {
		select {
		case connection := <-ch.register:
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*RedisTimeout)
			defer cancel()
			mMsg, err := ch.redis.GetAll(ctx)
			if err != nil {
				log.Println("websocket unable to dispatch old messages!")
				msg, err := ch.ErrorMessage("Dispatch SYSTEM: " + err.Error())
				connection.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					log.Println("websocket json marshal error message", string(err.Error()))
				}
				connection.Close()
				ch.unregister <- connection
				continue
			}
			msg, err := json.Marshal(mMsg)
			if err != nil {
				log.Println("websocket unable marshal and dispatch old messages!")
				msg, err := ch.ErrorMessage("Marshal Dispatch SYSTEM: " + err.Error())
				connection.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					log.Println("websocket json marshal error message", string(err.Error()))
				}
				connection.Close()
				ch.unregister <- connection
				continue
			}
			log.Printf("runHub return old messages: %s\n", msg)
			connection.WriteMessage(websocket.TextMessage, []byte(msg))
			ch.clients[connection] = &client{}

		case message := <-ch.broadcast:
			for connection, c := range ch.clients {
				go func(connection *websocket.Conn, c *client) {
					c.mu.Lock()
					defer c.mu.Unlock()
					if c.isClosing {
						return
					}
					log.Printf("runHub user message %s\n", message)
					if err := connection.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
						c.isClosing = true
						log.Println("write error:", err)
						connection.WriteMessage(websocket.CloseMessage, []byte{})
						connection.Close()
						ch.unregister <- connection
					}
				}(connection, c)
			}

		case connection := <-ch.unregister:
			delete(ch.clients, connection)
		}
	}
}

func (u *Chat) Register(c fiber.Router) {
	c.Get("/ws", u.ChatWebsocketAcceptorMiddleware, websocket.New(u.GetWebsocket))
	go u.runHub()
}
