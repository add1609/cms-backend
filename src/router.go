package src

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
	"strconv"

	"github.com/gorilla/websocket"
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

func parseFileTree(byteSlice []byte) map[string]interface{} {
	outSlice := bytes.Split(byteSlice, []byte("\n"))
	outMap := map[string]interface{}{}
	for i, element := range outSlice {
		if string(element) != "" {
			outMap[strconv.Itoa(i)] = string(element)
		}
	}
	delete(outMap, "0")
	return outMap
}

func (c *Client) handleRequest() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("[INFO] [%v] [id=%s] [pid=%v] %s", len(c.hub.clients), c.id, c.hugoPid, err.Error())
			break
		}
		var msgStr bytes.Buffer
		err = json.Compact(&msgStr, msg)
		if err != nil {
			log.Printf("[ERROR] [id=%s] [pid=%v] in handleRoutes(): %v", c.id, c.hugoPid, err)
			break
		}

		var reqMsg requestMessage
		err = json.Unmarshal(msg, &reqMsg)
		if err != nil {
			log.Printf("[ERROR] [id=%s] [pid=%v] in handleRoutes(): %v", c.id, c.hugoPid, err)
		} else if reqMsg.Action == "" || reqMsg.Payload == nil {
			log.Printf("[INFO] [%v] [id=%s] [pid=%v] BAD REQ: %v", len(c.hub.clients), c.id, c.hugoPid, &msgStr)
		} else {
			log.Printf("[INFO] [%v] [id=%s] [pid=%v] REQ: %s", len(c.hub.clients), c.id, c.hugoPid, &msgStr)
			switch reqMsg.Action {
			case "reqStartHugo":
				{
					if c.hugoPid == 0 {
						go c.startHugo()
					}
					c.resChan <- responseMessage{
						Action:  "resStartHugo",
						Success: true,
						Payload: map[string]interface{}{
							"previewUrl": c.url,
						},
					}
				}
			case "reqStopHugo":
				{
					if c.hugoPid != 0 {
						c.stopHugo()
					}
					c.resChan <- responseMessage{
						Action:  "resStopHugo",
						Success: true,
						Payload: map[string]interface{}{},
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
			case "reqAllFiles":
				{
					out, err := exec.Command("git", "submodule", "foreach", "git", "ls-files", "content/").Output()
					if err != nil {
						log.Printf("[ERROR] [id=%s] [pid=%v] in handleRoutes(): %v", c.id, c.hugoPid, err)
						return
					}
					outTree := parseFileTree(out)
					c.resChan <- responseMessage{
						Action:  "resAllFiles",
						Success: true,
						Payload: map[string]interface{}{
							"files": outTree,
						},
					}
				}
			case "reqModFiles":
				{
					out, err := exec.Command("git", "submodule", "foreach", "git", "ls-files", "-m", "content/").Output()
					if err != nil {
						log.Printf("[ERROR] [id=%s] [pid=%v] in handleRoutes(): %v", c.id, c.hugoPid, err)
						return
					}
					outTree := parseFileTree(out)
					c.resChan <- responseMessage{
						Action:  "resModFiles",
						Success: true,
						Payload: map[string]interface{}{
							"files": outTree,
						},
					}
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
	log.Printf("[INFO] [%v] [id=%s] [pid=%v] RES: %s", len(c.hub.clients), c.id, c.hugoPid, msg)
}
