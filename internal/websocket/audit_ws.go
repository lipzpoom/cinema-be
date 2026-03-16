package websocket

import (
	"log"
	"sync"

	"gin-quickstart/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type AuditLogClient struct {
	Conn *websocket.Conn
	Send chan models.AuditLog
}

type AuditLogManager struct {
	Clients    map[*AuditLogClient]bool
	Broadcast  chan models.AuditLog
	Register   chan *AuditLogClient
	Unregister chan *AuditLogClient
	mu         sync.Mutex
}

func NewAuditLogManager() *AuditLogManager {
	return &AuditLogManager{
		Clients:    make(map[*AuditLogClient]bool),
		Broadcast:  make(chan models.AuditLog),
		Register:   make(chan *AuditLogClient),
		Unregister: make(chan *AuditLogClient),
	}
}

func (manager *AuditLogManager) Start() {
	for {
		select {
		case client := <-manager.Register:
			manager.mu.Lock()
			manager.Clients[client] = true
			manager.mu.Unlock()

		case client := <-manager.Unregister:
			manager.mu.Lock()
			if _, ok := manager.Clients[client]; ok {
				delete(manager.Clients, client)
				close(client.Send)
			}
			manager.mu.Unlock()

		case message := <-manager.Broadcast:
			manager.mu.Lock()
			for client := range manager.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(manager.Clients, client)
				}
			}
			manager.mu.Unlock()
		}
	}
}

func (c *AuditLogClient) ReadPump() {
	defer func() {
		c.Conn.Close()
	}()
	// Keep reading to detect disconnections
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *AuditLogClient) WritePump(manager *AuditLogManager) {
	defer func() {
		manager.Unregister <- c
		c.Conn.Close()
	}()
	for {
		msg, ok := <-c.Send
		if !ok {
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		err := c.Conn.WriteJSON(msg)
		if err != nil {
			log.Printf("error writing JSON to audit ws: %v", err)
			break
		}
	}
}

func ServeAuditWS(manager *AuditLogManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("upgrade error:", err)
			return
		}

		client := &AuditLogClient{
			Conn: conn,
			Send: make(chan models.AuditLog, 256),
		}

		manager.Register <- client

		go client.WritePump(manager)
		go client.ReadPump()
	}
}