/*
 * 	Monitor specified topics.
 */

package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// this callback triggers when a message is received, it then prints the message (in the payload) and topic
var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	// fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	fmt.Printf("%s:%s\n", msg.Topic(), msg.Payload())
}

// upon connection to the client, this is called
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

// this is called when the connection to the client is lost, it prints "Connection lost" and the corresponding error
var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}

// disconnect at end of program
func disconnect(client mqtt.Client) {
	fmt.Print("disconnecting mqtt client\n")
	client.Disconnect(250)
}

// ConnectMqtt connects to MQTT allowing pub/sub
func ConnectMqtt() {
	// public secure server
	// var broker = "4bbda6ab78f84cdab69a7d5e9bfd21e7.s1.eu.hivemq.cloud" // find the host name in the Overview of your cluster (see readme)
	// var port = 8883

	_, mqttSecrets := readSecrets()

	var broker = mqttSecrets["server"]
	var port = int((mqttSecrets["port"]).(float64))

	// lan open server
	// var broker = "127.0.0.1"
	// broker = "10.0.0.241"
	// var port = 1883

	topics := []string{"#"}

	// public secure network

	/*
		opts := mqtt.NewClientOptions()
		opts.AddBroker(fmt.Sprintf("tls://%s:%d", broker, port))

			opts.SetClientID("mqtt_listener_v01") // set a name as you desire
			opts.SetUsername("ddgarrett")         // these are the credentials that you declare for your cluster (see readme)
			opts.SetPassword("x3PuYeAbznR3zt.")
	*/

	// private lan open servier

	var server = fmt.Sprintf("tcp://%s:%d", broker, port)
	fmt.Print("connecting to ")
	fmt.Println(server)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(server)
	opts.SetClientID("mqtt_listener_v01") // set a name as you desire

	if username, ok := mqttSecrets["username"]; ok {
		opts.SetUsername(username.(string))
		opts.SetPassword((mqttSecrets["password"]).(string))
	}

	// configure callback handlers
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	// create the client using the options above
	client := mqtt.NewClient(opts)
	// throw an error if the connection isn't successfull
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	defer disconnect(client)

	subscribe(client, topics)

	// loop forever, listening for subscribed events
	for {
		time.Sleep(time.Second * 10)
	}

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
