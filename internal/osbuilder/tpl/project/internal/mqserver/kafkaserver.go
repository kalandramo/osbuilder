package {{.MQ.Name}}

import (
	"context"

	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/onexstack/onexstack/pkg/server"

	{{- if eq .MQ.MQFramework "kafka"}}
	"{{.M.ModuleName}}/internal/pkg/kafka"
    {{- end}}
)

// mqServer implements the server.Server interface using the kafka framework.
type mqServer struct {
	engine *kafka.Engine
	obs    *genericoptions.ObservabilityOptions
}

// Ensure *mqServer implements the server.Server interface.
var _ server.Server = (*mqServer)(nil)

// NewMQServer initializes and returns a new HTTP server based on kafka.
// It returns the server.Server interface to abstract implementation details.
func (c *ServerConfig) NewMQServer() (server.Server, error) {
	kafkaClient, err := c.KafkaOptions.NewClient()
    if err != nil {
        return nil, err
    }
    engine, err := kafka.NewEngineFromClient(kafkaClient)
    if err != nil {
        return nil, err
    }

    c.Handler.ApplyTo(engine)

    return &mqServer{engine: engine, obs: c.ObservabilityOptions}, nil
}
// RunOrDie starts the Kafka server and panics if startup fails.
func (s *mqServer) RunOrDie(ctx context.Context) {
    go s.obs.Serve()
    s.engine.Start(ctx)
}

// GracefulStop gracefully shuts down the server.
func (s *mqServer) GracefulStop(ctx context.Context) {
	s.engine.Stop()
}
