package src

import (
	"encoding/json"
	"log"
)

type WsMessage struct {
	Action  string                 `json:"action"`
	Payload map[string]interface{} `json:"payload"`
}

func (c *Client) handleRoutes() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("[INFO] [id=%s] [pid=%v] in reader(): %v", c.id, c.hugoPid, err)
			break
		}
		var wsMsg WsMessage
		err = json.Unmarshal(msg, &wsMsg)
		if err != nil {
			log.Printf("[ERROR] [id=%s] [pid=%v] in handleRoutes(): %v", c.id, c.hugoPid, err)
		} else {
			log.Printf("[INFO] [id=%s] [pid=%v] Message recevied:\naction: %s, payload: %v",
				c.id, c.hugoPid, wsMsg.Action, wsMsg.Payload["message"])
		}
	}
}
