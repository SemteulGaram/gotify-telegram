package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gotify/plugin-api"
	"golang.org/x/net/websocket"
)

// GetGotifyPluginInfo returns gotify plugin info.
func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
		Version:    "1.0",
		Author:     "Kyush",
		Name:       "Gotify Telegram",
		License:    "MIT",
		ModulePath: "github.com/SemteulGaram/gotify-telegram",
	}
}

// MyPlugin is the gotify plugin instance.
type MyPlugin struct {
	ws                 *websocket.Conn
	gotify_url         string
	telegram_bot_token string
	telegram_chatid    string
}

type GotifyMessage struct {
	Id       uint32
	Appid    uint32
	Message  string
	Title    string
	Priority uint32
	Date     string
}

type Payload struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func (c *MyPlugin) get_websocket_msg() {
	go c.connect_websocket()

	for {
		msg := &GotifyMessage{}
		if c.ws == nil {
			time.Sleep(time.Second * 3)
			continue
		}
		// err := c.ws.ReadJSON(msg)
		err := websocket.JSON.Receive(c.ws, msg)
		if err != nil {
			fmt.Printf("Error while reading websocket: %v\n", err)
			c.connect_websocket()
			continue
		}
		c.send_msg_to_telegram(msg.Date + "\n" + msg.Title + "\n\n" + msg.Message)
	}
}

func (c *MyPlugin) connect_websocket() {
	for {
		// ws, _, err := websocket.DefaultDialer.Dial(p.gotify_host, nil)
		ws, err := websocket.Dial(c.gotify_url, "", "http://localhost/")
		if err == nil {
			c.ws = ws
			break
		}
		fmt.Printf("Cannot connect to websocket: %v\n", err)
		time.Sleep(time.Second * 5)
	}
}

func (p *MyPlugin) send_msg_to_telegram(msg string) {
	data := Payload{
		// Fill struct
		ChatID: p.telegram_chatid,
		Text:   msg,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Create json false")
		return
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://api.telegram.org/bot"+p.telegram_bot_token+"/sendMessage", body)
	if err != nil {
		fmt.Println("Create request false")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Send request false: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

// Enable enables the plugin.
func (c *MyPlugin) Enable() error {
	c.gotify_url = os.Getenv("GOTIFY_HOST") + "/stream?token=" + os.Getenv("GOTIFY_CLIENT_TOKEN")
	c.telegram_chatid = os.Getenv("TELEGRAM_CHAT_ID")
	c.telegram_bot_token = os.Getenv("TELEGRAM_BOT_TOKEN")
	go c.get_websocket_msg()
	return nil
}

// Disable disables the plugin.
func (c *MyPlugin) Disable() error {
	if c.ws != nil {
		c.ws.Close()
	}
	return nil
}

// RegisterWebhook implements plugin.Webhooker.
// func (c *MyPlugin) RegisterWebhook(basePath string, g *gin.RouterGroup) {
// }

// NewGotifyPluginInstance creates a plugin instance for a user context.
func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
	return &MyPlugin{}
}

func main() {
	panic("this should be built as go plugin")
}
