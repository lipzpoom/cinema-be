package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins for now
	},
}

type Client struct {
	Manager    *ClientManager
	Conn       *websocket.Conn
	Send       chan Message
	ShowtimeID string
}

func (c *Client) ReadPump() {
	defer func() {
		c.Manager.Unregister <- c
		c.Conn.Close()
	}()
	for {
		var msg Message
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("error reading JSON: %v", err)
			break
		}
		// ensure message has correct showtime
		msg.ShowtimeID = c.ShowtimeID
		c.Manager.Broadcast <- msg
	}
}

func (c *Client) WritePump() {
	defer func() {
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
			log.Printf("error writing JSON: %v", err)
			break
		}
	}
}

func ServeWS(manager *ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		showtimeID := c.Query("showtime_id")
		if showtimeID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "showtime_id is required"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("upgrade error:", err)
			return
		}

		client := &Client{
			Manager:    manager,
			Conn:       conn,
			Send:       make(chan Message, 256),
			ShowtimeID: showtimeID,
		}

		client.Manager.Register <- client

		go client.WritePump()
		go client.ReadPump()
	}
}
