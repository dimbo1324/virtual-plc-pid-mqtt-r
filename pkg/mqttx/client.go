package mqttx

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

// CommandHandler applies a parsed MQTT command to an application service.
type CommandHandler func(context.Context, plc.Command) (plc.Event, error)

// Client owns MQTT connection, publishing, and command subscription behavior.
type Client struct {
	config   Config
	handler  CommandHandler
	client   mqtt.Client
	deviceID string

	connectMu sync.Mutex
}

// New validates config and creates a client. Disabled clients are safe no-ops.
func New(config Config, handler CommandHandler) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	config = config.normalized()
	client := &Client{config: config, handler: handler, deviceID: deviceIDFromBaseTopic(config.BaseTopic)}
	if !config.Enabled {
		return client, nil
	}
	if handler == nil {
		return nil, fmt.Errorf("%w: command handler must not be nil", ErrInvalidConfig)
	}

	offline, err := MarshalStatus(StatusPayload{
		DeviceID: client.deviceID, Status: "offline", Reason: "unexpected_disconnect", Timestamp: time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal MQTT last will: %w", err)
	}

	options := mqtt.NewClientOptions().
		AddBroker(config.BrokerURL).
		SetClientID(config.ClientID).
		SetOrderMatters(false).
		SetAutoReconnect(true).
		SetConnectTimeout(config.ConnectTimeout).
		SetMaxReconnectInterval(config.ReconnectInterval).
		SetWill(StatusTopic(config.BaseTopic), string(offline), config.QoS, true)
	if config.Username != "" {
		options.SetUsername(config.Username)
	}
	if config.Password != "" {
		options.SetPassword(config.Password)
	}
	options.OnConnect = client.onConnect
	client.client = mqtt.NewClient(options)
	return client, nil
}

// Connect establishes the initial broker connection.
func (c *Client) Connect(ctx context.Context) error {
	if !c.config.Enabled {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	c.connectMu.Lock()
	defer c.connectMu.Unlock()
	if c.client.IsConnected() {
		return nil
	}
	token := c.client.Connect()
	if err := waitToken(ctx, token, c.config.ConnectTimeout); err != nil {
		return fmt.Errorf("connect MQTT broker %q: %w", c.config.BrokerURL, err)
	}
	return nil
}

// Disconnect publishes graceful offline status and closes the connection.
func (c *Client) Disconnect(quiesce uint) {
	if !c.config.Enabled || c.client == nil || !c.client.IsConnected() {
		return
	}
	status := StatusPayload{DeviceID: c.deviceID, Status: "offline", Reason: "graceful_shutdown", Timestamp: time.Now().UTC()}
	if payload, err := MarshalStatus(status); err == nil {
		token := c.client.Publish(StatusTopic(c.config.BaseTopic), c.config.QoS, true, payload)
		token.WaitTimeout(time.Second)
	}
	c.client.Disconnect(quiesce)
}

func (c *Client) PublishStatus(ctx context.Context, status StatusPayload) error {
	if !c.config.Enabled {
		return nil
	}
	payload, err := MarshalStatus(status)
	if err != nil {
		return fmt.Errorf("marshal MQTT status: %w", err)
	}
	return c.publish(ctx, StatusTopic(c.config.BaseTopic), payload, true)
}

func (c *Client) PublishSnapshot(ctx context.Context, snapshot plc.Snapshot) error {
	if !c.config.Enabled {
		return nil
	}
	payload, err := MarshalSnapshot(snapshot)
	if err != nil {
		return fmt.Errorf("marshal MQTT snapshot: %w", err)
	}
	return c.publish(ctx, TelemetryTopic(c.config.BaseTopic), payload, false)
}

func (c *Client) PublishEvent(ctx context.Context, event plc.Event) error {
	if !c.config.Enabled {
		return nil
	}
	payload, err := MarshalEvent(c.deviceID, event)
	if err != nil {
		return fmt.Errorf("marshal MQTT event: %w", err)
	}
	return c.publish(ctx, EventsTopic(c.config.BaseTopic), payload, false)
}

// IsConnected reports live broker connectivity.
func (c *Client) IsConnected() bool {
	return c.config.Enabled && c.client != nil && c.client.IsConnected()
}

func (c *Client) onConnect(client mqtt.Client) {
	token := client.Subscribe(CommandsTopic(c.config.BaseTopic), c.config.QoS, c.onMessage)
	if !token.WaitTimeout(c.config.ConnectTimeout) || token.Error() != nil {
		return
	}
	status := StatusPayload{DeviceID: c.deviceID, Status: "online", Timestamp: time.Now().UTC()}
	if payload, err := MarshalStatus(status); err == nil {
		client.Publish(StatusTopic(c.config.BaseTopic), c.config.QoS, true, payload)
	}
}

func (c *Client) onMessage(_ mqtt.Client, message mqtt.Message) {
	command, err := ParseCommand(message.Payload())
	if err != nil {
		c.publishRejected(commandIDFromPayload(message.Payload()), err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.config.ConnectTimeout)
	defer cancel()
	event, applyErr := c.handler(ctx, command)
	if applyErr != nil {
		if event.Type == "" {
			event = rejectedEvent(command.CommandID, applyErr)
		}
		_ = c.PublishEvent(ctx, event)
		return
	}
	if event.Type != "" {
		_ = c.PublishEvent(ctx, event)
	}
	applied := plc.Event{
		Timestamp: time.Now().UTC(), Level: "info", Type: plc.EventCommandApplied, Message: "command applied",
		Details: map[string]any{"command_id": command.CommandID, "command": command.Command, "event_type": event.Type},
	}
	_ = c.PublishEvent(ctx, applied)
}

func (c *Client) publishRejected(commandID string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.ConnectTimeout)
	defer cancel()
	_ = c.PublishEvent(ctx, rejectedEvent(commandID, err))
}

func rejectedEvent(commandID string, err error) plc.Event {
	details := map[string]any{}
	if commandID != "" {
		details["command_id"] = commandID
	}
	return plc.Event{Timestamp: time.Now().UTC(), Level: "warning", Type: plc.EventCommandRejected, Message: err.Error(), Details: details}
}

func (c *Client) publish(ctx context.Context, topic string, payload []byte, retained bool) error {
	if c.client == nil || !c.client.IsConnected() {
		return ErrNotConnected
	}
	token := c.client.Publish(topic, c.config.QoS, retained, payload)
	if err := waitToken(ctx, token, c.config.ConnectTimeout); err != nil {
		return fmt.Errorf("publish MQTT topic %q: %w", topic, err)
	}
	return nil
}

func waitToken(ctx context.Context, token mqtt.Token, fallback time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	timeout := fallback
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < timeout {
			timeout = remaining
		}
	}
	if timeout <= 0 {
		return context.DeadlineExceeded
	}
	if !token.WaitTimeout(timeout) {
		if err := ctx.Err(); err != nil {
			return err
		}
		return ErrTimeout
	}
	return token.Error()
}

func deviceIDFromBaseTopic(baseTopic string) string {
	parts := strings.Split(strings.Trim(baseTopic, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
