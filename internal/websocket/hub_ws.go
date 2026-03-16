package websocket

import "sync"

// Message represents the data sent through WebSocket
type Message struct {
	ShowtimeID string `json:"showtime_id"`
	SeatID     string `json:"seat_id"`
	Status     string `json:"status"` // E.g., "LOCKED", "UNLOCKED", "BOOKED"
}

// ClientManager maps showtime_id to connected clients
type ClientManager struct {
	Clients    map[string]map[*Client]bool
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		Clients:    make(map[string]map[*Client]bool),
		Broadcast:  make(chan Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (manager *ClientManager) Start() {
	for {
		select {
		case client := <-manager.Register:
			manager.mu.Lock()
			if manager.Clients[client.ShowtimeID] == nil {
				manager.Clients[client.ShowtimeID] = make(map[*Client]bool)
			}
			manager.Clients[client.ShowtimeID][client] = true
			manager.mu.Unlock()

		case client := <-manager.Unregister:
			manager.mu.Lock()
			if _, ok := manager.Clients[client.ShowtimeID][client]; ok {
				delete(manager.Clients[client.ShowtimeID], client)
				close(client.Send)
			}
			manager.mu.Unlock()

		case message := <-manager.Broadcast:
			manager.mu.Lock()
			// Send only to clients viewing the same showtime
			if clients, ok := manager.Clients[message.ShowtimeID]; ok {
				for client := range clients {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(manager.Clients[message.ShowtimeID], client)
					}
				}
			}
			manager.mu.Unlock()
		}
	}
}
