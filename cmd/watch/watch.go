package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"github.com/davecgh/go-spew/spew"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"google.golang.org/protobuf/proto"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	fmt.Println("hello world")

	var broker = "mqtt.meshtastic.org"
	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername("meshdev")
	opts.SetPassword("large4cats")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for _, t := range []string{
		"msh/+/#",
		"msh/US/2/e/#",
		"msh/US/+/2/e/#",
		"msh/US/+/+/2/e/#",
	} {
		topicToken := client.Subscribe(t, 0, MqttEventHandler)
		topicToken.Wait()
		if err := topicToken.Error(); err != nil {
			panic(err)
		}
		fmt.Printf("Subscribed to topic: %s\n", t)
	}

	// wait for ctl-c and then disconnect
	fmt.Println("Press CTRL+C to exit")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	fmt.Println("Disconnecting")

	client.Disconnect(250)

}

func MqttEventHandler(c mqtt.Client, m mqtt.Message) {
	if strings.Contains(m.Topic(), "/json/") {
		fmt.Println("Got a JSON message")
		envelope := meshtastic.ServiceEnvelope{}
		json.NewDecoder(bytes.NewReader(m.Payload())).Decode(&envelope)
		spew.Dump(envelope)
		return
	}
	envelope := meshtastic.ServiceEnvelope{}
	err := proto.Unmarshal(m.Payload(), &envelope)
	if err != nil {
		fmt.Println("couldn't decode wire packet: ", err)
		return
	}

	if envelope.Packet == nil {
		return
	}
	payloadBytes := make([]byte, 128)
	packetBody := envelope.Packet.GetDecoded()
	if packetBody == nil {
		if envelope.Packet.GetEncrypted() != nil {
			//fmt.Println("got an encrypted packet")
			return
			// TODO: decrypt the packet
			key := []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
			// decrypt the packet with aes256-ctr
			aesblock, err := aes.NewCipher(key)
			if err != nil {
				panic(err)
			}
			decryptor := cipher.NewCTR(aesblock, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

			// set up the output buffer:
			srcbuff := bufio.NewReader(bytes.NewReader(envelope.Packet.GetEncrypted()))
			dstbuff := bufio.NewWriter(bytes.NewBuffer(payloadBytes))

			reader := cipher.StreamReader{S: decryptor, R: srcbuff}
			io.Copy(dstbuff, reader)

			spew.Dump(payloadBytes)
			packetBodyx := meshtastic.Data{}
			if err := proto.Unmarshal(payloadBytes, &packetBodyx); err != nil {
				return
			}
			spew.Dump(packetBodyx)
			goto parse
		}
		return
	} else {
		payloadBytes = packetBody.GetPayload()
	}
	fmt.Printf("TOPIC: %s -> ", m.Topic())

parse:
	switch packetBody.Portnum {
	case meshtastic.PortNum_TEXT_MESSAGE_APP:
		fmt.Println("got a text message")
	case meshtastic.PortNum_POSITION_APP:
		fmt.Println("got a position report")
		payload := meshtastic.Position{}
		if err := proto.Unmarshal(payloadBytes, &payload); err != nil {
			fmt.Println("damn")
			panic(err)
		}
		spew.Dump(payload)
	case meshtastic.PortNum_NODEINFO_APP:
		fmt.Printf("got a node info from %x <> %x", packetBody.Dest, packetBody.Source)
		payload := meshtastic.NodeInfo{}
		proto.Unmarshal(payloadBytes, &payload)
		fmt.Printf("id: %x -> user %s with a %v\n", payload.User.Id, payload.User.LongName, payload.User.GetHwModel())
	case meshtastic.PortNum_TELEMETRY_APP:
		fmt.Println("got a telemetry packet")
		payload := meshtastic.Telemetry{}
		proto.Unmarshal(payloadBytes, &payload)
		spew.Dump(payload)

	default:
		fmt.Printf("got a packet of type %v\n", packetBody.Portnum)
		spew.Dump(packetBody)
	}
}
