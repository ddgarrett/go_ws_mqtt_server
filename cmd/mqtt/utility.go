/*
	Utility functions
*/

package main

import (
	"crypto/rand"
	"encoding/json"
	"os"
)

var fileName = "../../secrets.json"
var wifiEnv = "pixel"  // "home"  // "esp" // "pixel"
var mqttEnv = "hivemq" // "home_acer" // "hivemq" // "esp"

// getRandomClientId returns randomized ClientId.
func getRandomClientId() string {
	const alphanum = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var bytes = make([]byte, 12)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	// client id max length 24 in version <= V3.1
	return "GoWsMqtt" + string(bytes)
}

/*
func printDict(dict map[string]interface{}) {
	b, err := json.MarshalIndent(dict, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
*/

// GetSubdict returns a dictionary
// which is embedded within another dictionary
func GetSubdict(dict map[string]interface{}, subdictName string) map[string]interface{} {
	return dict[subdictName].(map[string]interface{})
}

func readSecrets() (wifi map[string]interface{}, mqtt map[string]interface{}) {
	fileBytes, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(fileBytes, &data)
	if err != nil {
		panic(err)
	}

	// fmt.Println("wifi: ")
	// printDict(GetSubdict(GetSubdict(data, "wifi"), wifiEnv))

	// fmt.Println("mqtt: ")
	// printDict(GetSubdict(GetSubdict(data, "mqtt"), mqttEnv))

	return GetSubdict(GetSubdict(data, "wifi"), wifiEnv),
		GetSubdict(GetSubdict(data, "mqtt"), mqttEnv)
}
