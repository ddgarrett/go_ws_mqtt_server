# Go Websocket MQTT Server

Connects to an MQTT broker from a web browser via Websockets.

Allows subscribe and publish calls.

Uses [Paho MQTT for Go](github.com/eclipse/paho.mqtt.golang) and [Gorrilla Websockets](github.com/gorilla/websocket).

To use, execute using the following:

```.../cmd/mqtt/.exe```

This will start a server running on port 8080 and use the MQTT broker "hivemq". The broker, "hivemq", must be defined in a secrets.json file. See example below. 

The port and MQTT broker may be overridden:

```.../cmd/mqtt/mqtt.exe -addr ":8081" -mqtt "rpi400_mqtt"```

This example would start a server running on port 8081, accessing the MQTT broker defined in "rpi400_mqtt" in the example `secrets.json` file below.

Note that in this example `secrets.json` file the HiveMQ server name, username and password have been redacted.

```json
{
    "mqtt" :  {
        "rpi400_mqtt" :{
            "server" : "10.0.0.231",
            "port"   : 1883
            },    
        "hivemq" :{
            "server" : "<hivemq server prefix>.s1.eu.hivemq.cloud",
            "port"   : 8883,
            "username" : "<hivemq username>",
            "password" : "<hivemq password>",
            "ca_cert"  : "hivemq_root_ca.dat"
            }
    }
}
```

On a browser on the same machine open a browser window and go to the localhost. For the two examples above this would be 

 ```http://localhost:8080``` 

or
```http://localhost:8081```


TODO: add screen capture of web page, and sub/pub commands