package client

import (
	"fmt"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"google.golang.org/protobuf/proto"
)

type MeshtasticClient struct {
	client   mqtt.Client
	Messages chan meshtastic.ServiceEnvelope
}

func NewClient(config *mqtt.ClientOptions, prefix string) *MeshtasticClient {
	mqttClient := mqtt.NewClient(config)
	token := mqttClient.Connect()
	token.WaitTimeout(time.Second * 10)
	if token.Error() != nil {
		er := token.Error()
		fmt.Printf("%v \n", er)
		panic(token.Error())
	}

	mclient := &MeshtasticClient{
		client:   mqttClient,
		Messages: make(chan meshtastic.ServiceEnvelope),
	}

	mqttClient.Subscribe(prefix, 0, func(client mqtt.Client, msg mqtt.Message) {
		packet := meshtastic.ServiceEnvelope{}
		err := proto.Unmarshal(msg.Payload(), &packet)
		if err != nil {
			panic(err)
		}
		mclient.Messages <- packet
	})

	return mclient
}
