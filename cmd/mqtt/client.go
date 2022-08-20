// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// MQTT info
	mqttClient mqtt.Client
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		c.mqttClient.Disconnect(250)
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.mqttClient.Disconnect(250)
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
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

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn,
		send: make(chan []byte, 256), mqttClient: nil}

	client.connect_mqtt()
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

/*
	MQTT Additions

*/

func (c *Client) connect_mqtt() {
	_, mqttSecrets := readSecrets()
	broker := mqttSecrets["server"].(string)
	port := int((mqttSecrets["port"]).(float64))
	topics := []string{"#"}

	opts := mqtt.NewClientOptions()

	if username, ok := mqttSecrets["username"]; ok {
		// tls protocol - encrypted
		opts.AddBroker(fmt.Sprintf("tls://%s:%d", broker, port))
		opts.SetUsername(username.(string))
		opts.SetPassword((mqttSecrets["password"]).(string))
	} else {
		// tcp protocol - unencrypted
		opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	}

	cid := getRandomClientId()
	fmt.Printf("using client id: %s\n", cid)
	opts.SetClientID(cid) // set a name as you desire

	// configure callback handlers
	opts.SetDefaultPublishHandler(c.messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	// create the client using the options above
	c.mqttClient = mqtt.NewClient(opts)

	// throw an error if the connection isn't successfull
	if token := c.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	c.subscribe(topics)
}

func (c *Client) messagePubHandler(client mqtt.Client, msg mqtt.Message) {
	message := fmt.Sprintf("rcv %s %s\n", msg.Topic(), msg.Payload())
	fmt.Println(message)
	c.send <- []byte(message)
}

// upon connection to the client, this is called
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected ")
}

// this is called when the connection to the client is lost, it prints "Connection lost" and the corresponding error
var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}

func (c *Client) subscribe(topics []string) {
	// subscribe list of topics
	for _, topic := range topics {
		token := c.mqttClient.Subscribe(topic, 1, nil)
		token.Wait()
		// Check for errors during subscribe
		if token.Error() != nil {
			fmt.Printf("Failed to subscribe to topic %s \n", topic)
			panic(token.Error())
		}
		fmt.Printf("Subscribed to topic: %s \n", topic)
	}

}
