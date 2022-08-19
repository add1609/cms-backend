package src

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

type requestMessage struct {
	Action  string                 `json:"action"`
	Payload map[string]interface{} `json:"payload"`
}
type responseMessage struct {
	Action  string                 `json:"action"`
	Success bool                   `json:"success"`
	Payload map[string]interface{} `json:"payload"`
}

func (c *Client) handleRequest() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("[%v] [INFO] [id=%s] [pid=%v] in handleRoutes(): %s", len(c.hub.clients), c.id, c.hugoPid, err.Error())
			break
		}
		var msgStr bytes.Buffer
		err = json.Compact(&msgStr, msg)
		if err != nil {
			log.Printf("[ERROR] [id=%s] [pid=%v] in handleRoutes(): %v", c.id, c.hugoPid, err)
			break
		}
		log.Printf("[%v] [INFO] [id=%s] [pid=%v] REQ: %s", len(c.hub.clients), c.id, c.hugoPid, &msgStr)

		var reqMsg requestMessage
		err = json.Unmarshal(msg, &reqMsg)
		if err != nil {
			log.Printf("[ERROR] [id=%s] [pid=%v] in handleRoutes(): %v", c.id, c.hugoPid, err)
			break
		}
		switch reqMsg.Action {
		case "reqPreviewUrl":
			{
				c.resChan <- responseMessage{
					Action:  "resPreviewUrl",
					Success: true,
					Payload: map[string]interface{}{
						"previewUrl": c.url,
					},
				}
			}
		case "reqClientId":
			{
				c.resChan <- responseMessage{
					Action:  "resClientId",
					Success: true,
					Payload: map[string]interface{}{
						"id": c.id,
					},
				}
			}
		}
	}
}

func (c *Client) handleResponse(resMsg responseMessage) {
	msg, err := json.Marshal(resMsg)
	if err != nil {
		log.Printf("[ERROR] [id=%s] [pid=%v] in handleSend(): %v", c.id, c.hugoPid, err)
		return
	}

	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		log.Printf("[ERROR] [id=%s] [pid=%v] in handleSend(): %v", c.id, c.hugoPid, err)
		return
	}

	w.Write(msg)
	if err = w.Close(); err != nil {
		log.Printf("[ERROR] [id=%s] [pid=%v] in handleSend(): %v", c.id, c.hugoPid, err)
		return
	}
	log.Printf("[%v] [INFO] [id=%s] [pid=%v] RES: %s", len(c.hub.clients), c.id, c.hugoPid, msg)
}
