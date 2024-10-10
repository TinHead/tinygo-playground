package main

import (
	// "errors"
	"encoding/json"
	"fmt"
	// "io"
	"log/slog"
	"os"
	"time"

	"github.com/soypat/cyw43439"
	// "golang.org/x/tools/present"
)

func main() {
	// Wait for USB to initialize:
	time.Sleep(time.Second)
	dev := cyw43439.NewPicoWDevice()
	cfg := cyw43439.DefaultWifiConfig()
	// cfg.Logger = logger // Uncomment to see in depth info on wifi device functioning.
	err := dev.Init(cfg)

	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.Level(-4), // Make temporary logger that does no logging.
		}))
	}
	if err != nil {
		panic(err)
	}

	stack := initWifi(dev, logger)
	broker := "192.168.1.3:1883"
	// Begin asynchronous packet handling.
	// Get a transport for MQTT packets
	fmt.Printf("Connecting to MQTT broker at %s\n", broker)
	// Start TCP server.

	conn, rng := setupClient(stack, logger, broker)
	client, err := initMqtt(conn, logger)
	if err != nil {
		logger.Info("Failed to connect to mqtt", err)
	}
	gw_present := presentPayload{
		Name:          "rfm-gw",
		DeviceClass:   "running",
		StateTopic:    "homeassistant/binary_sensor/rfm-gw/state",
		CommandTopic:  "homeassistant/binary_sensor/rfm-gw/set",
		UniqueId:      "gw01",
		PayloadOn:     "on",
		PayloadOff:    "off",
		ValueTemplate: "{{ value_json.state }}",
		Device: deviceStruct{
			Identifiers: []string{"gw01"},
			Name:        "gw01",
		},
	}

	payload, err := json.Marshal(gw_present)
	if err != nil {
		logger.Info(err.Error())
	} else {
		logger.Info(string(payload))
	}

	err = pubMqtt(client, conn, logger, rng, "homeassistant/binary_sensor/rfm-gw/config", payload)
	subMqtt(client, logger, rng, gw_present.CommandTopic)
	// publish my info to hass
	// err = pubMqtt(client, conn, logger, rng, "homeassistant/binary_sensor/rfm-gw/config", "'name': 'upy-gw','device_class': 'running', 'state_topic': 'homeassistant/binary_sensor/upy-gw/state',        'command_topic': 'homeassistant/binary_sensor/upy-gw/set','unique_id': 'upy-gw','device': {'identifiers': [\"gw01\"], 'name': \"upy_gw01\"}})")
	// blink led
	for {
		// err := pubMqtt(client, conn, logger, rng, "homeassistant", "hello")
		err = dev.GPIOSet(0, true)
		if err != nil {
			println("err", err.Error())
		} else {
			println("LED ON")
		}

		err = pubMqtt(client, conn, logger, rng, "homeassistant/binary_sensor/rfm-gw/state", []byte("{ \"state\": \"on\" }"))
		time.Sleep(2000 * time.Millisecond)
		err = dev.GPIOSet(0, false)
		if err != nil {
			println("err", err.Error())
		} else {
			println("LED OFF")
		}
		time.Sleep(2000 * time.Millisecond)
		err = pubMqtt(client, conn, logger, rng, "homeassistant/binary_sensor/rfm-gw/state", []byte("{ \"state\": \"off\" }"))
	}
}
