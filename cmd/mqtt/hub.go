// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// MQTT info
	mqttClient mqtt.Client
}

func newHub() *Hub {

	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		mqttClient: nil,
	}
}

func (h *Hub) run() {
	if h.mqttClient == nil {
		h.connect_mqtt()
		defer h.disconnect()
	}

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
func (h *Hub) connect_mqtt() {
	_, mqttSecrets := readSecrets()
	broker := mqttSecrets["server"].(string)
	port := int((mqttSecrets["port"]).(float64))
	topics := []string{"#"}

	server := ""
	opts := mqtt.NewClientOptions()

	if username, ok := mqttSecrets["username"]; ok {
		opts.SetUsername(username.(string))
		opts.SetPassword((mqttSecrets["password"]).(string))
		server = fmt.Sprintf("tls://%s:%d", broker, port)
	} else {
		server = fmt.Sprintf("tcp://%s:%d", broker, port)
	}

	opts.AddBroker(server)
	opts.SetClientID("go_ws_mqtt_v01") // set a name as you desire

	// configure callback handlers
	opts.SetDefaultPublishHandler(h.messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	// create the client using the options above
	h.mqttClient = mqtt.NewClient(opts)
	// throw an error if the connection isn't successfull
	if token := h.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// defer disconnect(client)

	subscribe(h.mqttClient, topics)
}

func (h *Hub) disconnect() {
	fmt.Print("disconnecting mqtt client\n")
	h.mqttClient.Disconnect(250)
}

// this callback triggers when a message is received, it then prints the message (in the payload) and topic
/*
var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	// fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	fmt.Printf("%s:%s\n", msg.Topic(), msg.Payload())
}
*/

func (h *Hub) messagePubHandler(client mqtt.Client, msg mqtt.Message) {
	// fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	// fmt.Printf("%s:%s\n", msg.Topic(), msg.Payload())

	message := fmt.Sprintf("rcv %s %s\n", msg.Topic(), msg.Payload())
	fmt.Println(message)

	h.broadcast <- []byte(message)
	/*
		for client := range h.clients {
			select {
			case client.send <- []byte(message):
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	*/
}

// upon connection to the client, this is called
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

// this is called when the connection to the client is lost, it prints "Connection lost" and the corresponding error
var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}

func subscribe(client mqtt.Client, topics []string) {
	// subscribe list of topics
	for _, topic := range topics {
		token := client.Subscribe(topic, 1, nil)
		token.Wait()
		// Check for errors during subscribe
		if token.Error() != nil {
			fmt.Printf("Failed to subscribe to topic %s \n", topic)
			panic(token.Error())
		}
		fmt.Printf("Subscribed to topic: %s \n", topic)
	}

}
