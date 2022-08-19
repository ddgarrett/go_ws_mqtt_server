/*
	Test json parsing logic
*/

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var fileName = "../../secrets.json"
var wifiEnv = "wifi_esp"
var mqttEnv = "mqtt_docker_esp"

func printDict(dict map[string]interface{}) {
	b, err := json.MarshalIndent(dict, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}

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

	/*
		fmt.Println("wifi: ")
		printDict(GetSubdict(GetSubdict(data, "wifi"), wifiEnv))

		fmt.Println("mqtt: ")
		printDict(GetSubdict(GetSubdict(data, "mqtt"), mqttEnv))
	*/

	return GetSubdict(GetSubdict(data, "wifi"), wifiEnv),
		GetSubdict(GetSubdict(data, "mqtt"), mqttEnv)
}
