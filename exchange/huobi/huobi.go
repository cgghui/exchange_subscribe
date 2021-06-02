package huobi

import (
	"encoding/json"
	"fmt"
	"github.com/cgghui/exchange_subscribe/exchange"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const DefaultWebsocketURL = "wss://api.huobi.pro/ws/v2"

type Connect struct {
	domain string
	path   string
	ak     string
	sk     string
	mutex  *sync.Mutex
	conn   *websocket.Conn
}

func NewConnect(ak, sk string) (*Connect, error) {
	conn, _, err := websocket.DefaultDialer.Dial(DefaultWebsocketURL, nil)
	if err != nil {
		return nil, err
	}
	obj := &Connect{"api.huobi.pro", "/ws/v2", ak, sk, &sync.Mutex{}, conn}
	go func() {
		defer obj.Close()
		for {
			var message []byte
			_, message, err = conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("%s\n", message)
			go obj.handleMessage(message)
		}
	}()
	prams := map[string]string{
		"authType":         "api",
		"accessKey":        ak,
		"signatureMethod":  "HmacSHA256",
		"signatureVersion": "2.1",
	}
	body := map[string]interface{}{"action": "req", "ch": "auth", "params": prams}
	if err = obj.buildParams(&prams); err != nil {
		return nil, err
	}
	_ = conn.WriteJSON(body)
	return obj, nil
}

func (c *Connect) handleMessage(body []byte) {
	var data MessageBase
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("huo bi parse message error: %s", err.Error())
		return
	}
	if data.Action == "ping" {
		go c.pong(body)
	}
}

func (c *Connect) pong(body []byte) {
	var ret MessagePing
	_ = json.Unmarshal(body, &ret)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	payload := MessagePing{MessageBase{Action: "pong"}, MessagePingBody{TS: ret.Data.TS}}
	data, _ := json.Marshal(&payload)
	log.Printf("%s\n", data)
	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("huo bi pong error: %s", err.Error())
	}
}

func (c *Connect) buildParams(data *map[string]string) error {
	(*data)["timestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05")
	var keys []string
	for k, _ := range *data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var form []string
	for _, k := range keys {
		form = append(form, fmt.Sprintf("%s=%s", k, (*data)[k]))
	}
	payload := fmt.Sprintf("%s\n%s\n%s\n%s", http.MethodGet, c.domain, c.path, strings.Join(form, "&"))
	fmt.Printf("%s\n", payload)
	sign, err := exchange.GetParamHmacSHA256Base64Sign(c.sk, payload)
	if err != nil {
		return err
	}
	(*data)["signature"] = sign
	return nil
}

func (c *Connect) Close() {
	err := c.conn.Close()
	if err != nil {
		log.Printf("huo bi websocket close error: %s", err.Error())
	}
}
