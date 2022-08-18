package src

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

// The Client is a middleman between the websocket connection and the Hub.
type Client struct {
	hub *Hub

	// The websocket connection of this Client.
	conn *websocket.Conn

	// The Client's buffered channel of outbound messages.
	send chan []byte

	// The Client's id.
	id string

	// The url on which the hugo server hosts this Client's preview.
	url string

	// The pid of the "hugo server" command, which is called by
	// startHugo to host this Client's preview.
	hugoPid int
}

// genId sets the Client's id field to a value between 2000 and 9999.
func (c *Client) genId() {
	deca64, _ := strconv.ParseInt(strings.Replace(fmt.Sprintf("%p", c)[6:], "0x", "", -1), 16, 64)
	deca := (int(deca64) % 7999) + 2000
	c.id = strconv.Itoa(deca)
}

// From frontend to Client.
func (c *Client) reader() {
	defer func() {
		c.stopHugo()
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(int64(upgrader.ReadBufferSize))
	c.conn.SetReadDeadline(time.Now().Add(c.hub.pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(c.hub.pongWait)); return nil })
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("[INFO] [id=%s] [pid=%v] in reader(): %v", c.id, c.hugoPid, err)
			break
		}
		log.Printf("[INFO] [id=%s] [pid=%v] Message recevied: %s", c.id, c.hugoPid, string(msg))
	}
}

// From Client to Frontend.
func (c *Client) writePump() {
	pingTicker := time.NewTicker(c.hub.pingPeriod)
	defer func() {
		pingTicker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-pingTicker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *Client) startHugo() {
	cmd := exec.Command("hugo", "server", "--baseURL", c.url, "--bind",
		c.hub.env["HUGO_BIND"], "--port", c.id, "--source", c.hub.env["HUGO_SOURCE"])

	if err := cmd.Start(); err != nil {
		log.Printf("[ERROR] [id=%s] [pid=%v] in startHugo(): %v", c.id, c.hugoPid, err)
	}
	c.hugoPid = cmd.Process.Pid
	log.Printf("[INFO] [id=%s] [pid=%v] Starting hugo server", c.id, c.hugoPid)
	if err := cmd.Wait(); err != nil {
		log.Printf("[ERROR] [id=%s] [pid=%v] in startHugo(): %v", c.id, c.hugoPid, err)
	}
}

func (c *Client) stopHugo() {
	if c.hugoPid != 0 {
		log.Printf("[INFO] [id=%s] [pid=%v] Stopping hugo server", c.id, c.hugoPid)
		cmd := exec.Command("kill", "-s", "INT", strconv.Itoa(c.hugoPid))
		if err := cmd.Run(); err != nil {
			log.Printf("[ERROR] [id=%s] [pid=%v] in stopHugo(): %v", c.id, c.hugoPid, err)
		}
		c.hugoPid = 0
	}
}

// setupWs sets up some parameters whenever ServeWs is called.
func setupWs(envReadSize, envWriteSize, envCheckOrigin, envOriginHost string) {
	rBufferSize, err := strconv.Atoi(envReadSize)
	if err != nil {
		rBufferSize = 1024
	}
	wBufferSize, err := strconv.Atoi(envWriteSize)
	if err != nil {
		wBufferSize = 1024
	}
	upgrader.ReadBufferSize = rBufferSize
	upgrader.WriteBufferSize = wBufferSize
	upgrader.CheckOrigin = func(req *http.Request) bool {
		if originHost == "" {
			return true
		}
		origin := req.Header["Origin"]
		if len(origin) == 0 {
			return false
		}
		u, err := url.Parse(origin[0])
		if err != nil {
			return false
		}
		return u.Host == envOriginHost
	}
}

// ServeWs handles websocket requests from the frontend.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	setupWs(
		hub.env["WS_READ_BUFFER_SIZE"],
		hub.env["WS_WRITE_BUFFER_SIZE"],
		hub.env["WS_CHECK_ORIGIN"],
		hub.env["WS_CHECK_ORIGIN_HOST"],
	)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ERROR] in upgrader.Upgrade(): %v", err)
		return
	}
	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	client.genId()
	client.url = hub.env["HUGO_BASE_URL"] + client.id + "/preview/"
	client.hub.register <- client

	go client.reader()
	go client.writePump()
	go client.startHugo()
}
