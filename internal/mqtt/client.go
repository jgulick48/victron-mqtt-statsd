package mqtt

import (
	"fmt"
	"github.com/jgulick48/victron-mqtt-statsd/internal/vebus"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client interface {
	Close()
	Connect()
}

type Configuration struct {
	Host     string
	Port     int
	DeviceID string
}

func NewClient(config Configuration) Client {
	client := client{
		config: config,
		done:   make(chan bool),
	}
	return &client
}

type client struct {
	config     Configuration
	done       chan bool
	mqttClient mqtt.Client
}

func (c *client) Close() {
	c.done <- true
}

func (c *client) Connect() {
	fmt.Printf("Connecting to %s", fmt.Sprintf("tcp://%s:%d", c.config.Host, c.config.Port))
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", c.config.Host, c.config.Port))
	opts.SetClientID("go_mqtt_client")
	opts.SetDefaultPublishHandler(c.messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = c.connectLostHandler
	c.mqttClient = mqtt.NewClient(opts)
	if token := c.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	c.sub()
	defer c.mqttClient.Disconnect(250)
	c.keepAlive()
}

func (c *client) keepAlive() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			token := c.mqttClient.Publish(fmt.Sprintf("R/%s/system/0/Serial", c.config.DeviceID), 0, false, "")
			token.Wait()
		}
	}
}

func (c *client) messagePubHandler(client mqtt.Client, msg mqtt.Message) {
	vebus.ProcessData(msg.Topic(), msg.Payload())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

func (c *client) connectLostHandler(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
	c.done <- true
}

func (c *client) sub() {
	topic := "#"
	token := c.mqttClient.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s", topic)
}
