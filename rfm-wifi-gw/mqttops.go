package main

import (
	// "errors"
	// "context"
	mqtt "github.com/soypat/natiu-mqtt"
	"github.com/soypat/seqs/stacks"
	"io"
	"log/slog"
	"math/rand"
	"time"
)

type deviceStruct struct {
	Identifiers []string `json:"identifiers"`
	Name        string   `json:"name"`
}

type presentPayload struct {
	Name          string       `json:"name"`
	DeviceClass   string       `json:"device_class"`
	StateTopic    string       `json:"state_topic"`
	CommandTopic  string       `json:"command_topic"`
	UniqueId      string       `json:"unique_id"`
	Device        deviceStruct `json:"device"`
	PayloadOn     string       `json:"payload_on"`
	PayloadOff    string       `json:"payload_off"`
	ValueTemplate string       `json:"value_template"`
}

type statePayload struct {
	State string `json:"state"`
}

func initMqtt(conn *stacks.TCPConn, logger *slog.Logger) (*mqtt.Client, error) {
	mqttcfg := mqtt.ClientConfig{
		Decoder: mqtt.DecoderNoAlloc{UserBuffer: make([]byte, 4096)},
		OnPub: func(pubHead mqtt.Header, varPub mqtt.VariablesPublish, r io.Reader) error {
			logger.Info("received message", slog.String("topic", string(varPub.TopicName)))
			return nil
		},
	}

	var varconn mqtt.VariablesConnect
	varconn.SetDefaultMQTT([]byte("rfm-gw"))
	varconn.Username = []byte(mqtt_user)
	varconn.Password = []byte(mqtt_pass)
	client := mqtt.NewClient(mqttcfg)
	logger.Info("mqtt:start-connecting")
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	err := client.StartConnect(conn, &varconn)
	if err != nil {
		logger.Error("mqtt:start-connect-failed", slog.String("reason", err.Error()))
		closeConn(conn, "connect failed")
	}
	retries := 50
	for retries > 0 && !client.IsConnected() {
		time.Sleep(100 * time.Millisecond)
		err = client.HandleNext()
		if err != nil {
			println("mqtt:handle-next-failed", err.Error())
		}
		retries--
	}
	logger.Info("mqtt:isconn? %t", client.IsConnected())
	return client, err
}

func pubMqtt(client *mqtt.Client, conn *stacks.TCPConn, logger *slog.Logger, rng *rand.Rand, topic string, msg []byte) error {

	pubFlags, _ := mqtt.NewPublishFlags(mqtt.QoS0, false, false)
	pubVar := mqtt.VariablesPublish{
		TopicName:        []byte(topic),
		PacketIdentifier: 0xc0fe,
	}

	// if client.IsConnected() {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	pubVar.PacketIdentifier = uint16(rng.Uint32())
	err := client.PublishPayload(pubFlags, pubVar, msg)
	if err != nil {
		logger.Error("mqtt:publish-failed", slog.Any("reason", err))
		return err
	}
	logger.Info("published message", slog.Uint64("packetID", uint64(pubVar.PacketIdentifier)))
	err = client.HandleNext()
	if err != nil {
		println("mqtt:handle-next-failed", err.Error())
		return err
	}
	// time.Sleep(5 * time.Second)
	// }
	return nil //errors.New("Mqtt connection down!")
}

func subMqtt(client *mqtt.Client, logger *slog.Logger, rng *rand.Rand, topic string) {
	vsub := mqtt.VariablesSubscribe{
		TopicFilters: []mqtt.SubscribeRequest{
			{TopicFilter: []byte(topic), QoS: mqtt.QoS0}, // Only support QoS0 for now.
		},
		PacketIdentifier: uint16(rng.Int31()),
	}
	// ctx := context.Background()
	err := client.StartSubscribe(vsub)
	if err != nil {
		logger.Error("could not subscribe, ", err.Error())
	}

	err = client.HandleNext()
	if err != nil {
		println("mqtt:handle-next-failed", err.Error())
	}
}
