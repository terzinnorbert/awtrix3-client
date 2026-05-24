package api

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTClient wraps the Eclipse Paho MQTT client for AWTRIX topic publishing.
type MQTTClient struct {
	client mqtt.Client
	prefix string
}

// NewMQTTClient creates and connects a new MQTTClient.
// brokerURL example: "tcp://192.168.1.10:1883"
// prefix is the AWTRIX MQTT prefix (e.g. "awtrix").
func NewMQTTClient(brokerURL, prefix, clientID string) (*MQTTClient, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("MQTT connect: %w", token.Error())
	}
	return &MQTTClient{client: c, prefix: prefix}, nil
}

// Publish JSON-encodes payload and publishes it to [prefix]/[topic].
// A nil payload publishes an empty message (for clear/delete operations).
func (m *MQTTClient) Publish(topic string, payload interface{}) error {
	var data []byte
	var err error
	if payload != nil {
		data, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}
	token := m.client.Publish(m.prefix+"/"+topic, 0, false, data)
	token.Wait()
	return token.Error()
}

// PublishRaw publishes a raw string payload (used for the rtttl topic).
func (m *MQTTClient) PublishRaw(topic string, payload string) error {
	token := m.client.Publish(m.prefix+"/"+topic, 0, false, payload)
	token.Wait()
	return token.Error()
}

// Disconnect cleanly disconnects the MQTT client.
func (m *MQTTClient) Disconnect() {
	m.client.Disconnect(250)
}
