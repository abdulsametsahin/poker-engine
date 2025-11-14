package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	UserID  string
	TableID string
	Conn    *websocket.Conn
	Send    chan []byte
}

// ReadPump handles incoming messages from the client
// CRITICAL: Mutex protection added to prevent concurrent map access panics
func (c *Client) ReadPump(clients map[string]interface{}, mu *sync.RWMutex, handleMessage func(*Client, WSMessage)) {
	defer func() {
		// CRITICAL: Protect map deletion with mutex to prevent server crashes
		mu.Lock()
		delete(clients, c.UserID)
		mu.Unlock()
		c.Conn.Close()
	}()

	for {
		var msg WSMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		handleMessage(c, msg)
	}
}

// WritePump handles outgoing messages to the client
func (c *Client) WritePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// GetTableID returns the table ID the client is subscribed to
func (c *Client) GetTableID() string {
	return c.TableID
}

// GetSendChannel returns the send channel for this client
func (c *Client) GetSendChannel() chan []byte {
	return c.Send
}
