package websocket

import (
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
func (c *Client) ReadPump(clients map[string]interface{}, handleMessage func(*Client, WSMessage)) {
	defer func() {
		delete(clients, c.UserID)
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
