package handler

import (
	"log"
	"sync"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/store"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// Add more data to this type if needed
type client struct {
	isClosing bool
	mu        sync.Mutex
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
			ch.broadcast <- string(message)
		} else {
			log.Println("websocket message received of type", messageType)
		}
	}
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
