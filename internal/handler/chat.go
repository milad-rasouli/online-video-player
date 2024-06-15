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
	clients    map[*websocket.Conn]*client // Note: although large maps with pointer-like types (e.g. strings) as keys are slow, using pointers themselves as keys is acceptable and fast
	register   chan *websocket.Conn
	broadcast  chan string
	unregister chan *websocket.Conn
	cfg        config.Config
	redis      store.MessageStore
}

func NewChat(cfg config.Config, redis store.MessageStore) *Chat {
	var clients = make(map[*websocket.Conn]*client) // Note: although large maps with pointer-like types (e.g. strings) as keys are slow, using pointers themselves as keys is acceptable and fast
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
	// When the function returns, unregister the client and close the connection
	defer func() {
		ch.unregister <- c
		c.Close()
	}()

	// Register the client
	ch.register <- c

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}

			return // Calls the deferred function, i.e. closes the connection on error
		}

		if messageType == websocket.TextMessage {
			// Broadcast the received message
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
			//TODO: store in redis
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

// this should be shutdown properly. in the case i mean
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
			connection.WriteMessage(websocket.TextMessage, []byte(msg))
			ch.clients[connection] = &client{}
			log.Println("connection registered")

		case message := <-ch.broadcast:
			log.Println("message received:", message)
			// Send the message to all clients
			for connection, c := range ch.clients {
				go func(connection *websocket.Conn, c *client) { // send to each client in parallel so we don't block on a slow client
					c.mu.Lock()
					defer c.mu.Unlock()
					if c.isClosing {
						return
					}
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
			// Remove the client from the hub
			delete(ch.clients, connection)
			log.Println("connection unregistered")
		}
	}
}

func GetChatPage(c *fiber.Ctx) error {
	return c.Render("chat", fiber.Map{})
}
func (u *Chat) Register(c fiber.Router) {
	c.Get("/ws", u.ChatWebsocketAcceptorMiddleware, websocket.New(u.GetWebsocket))
	c.Get("/page", GetChatPage)
	go u.runHub()

}
