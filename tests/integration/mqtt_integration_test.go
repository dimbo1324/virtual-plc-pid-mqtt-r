package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/mqttx"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

func TestMQTTConnection(t *testing.T) {
	if os.Getenv("VPLC_RUN_MQTT_TESTS") != "1" {
		t.Skip("set VPLC_RUN_MQTT_TESTS=1 to run broker integration")
	}
	client, err := mqttx.New(mqttx.Config{
		Enabled: true, BrokerURL: "tcp://localhost:1883", ClientID: "vplc-integration-test",
		BaseTopic: "vplc/integration-test", QoS: 0, ConnectTimeout: 3 * time.Second, ReconnectInterval: time.Second,
	}, func(context.Context, plc.Command) (plc.Event, error) { return plc.Event{}, nil })
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Disconnect(250)
	if !client.IsConnected() {
		t.Fatal("client is not connected")
	}
}
