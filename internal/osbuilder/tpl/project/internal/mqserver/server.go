package {{.MQ.Name}}

import (
	"context"
	"log/slog"
    "time"

    {{- if eq .MQ.StorageType "memory" }}
	"github.com/onexstack/onexstack/pkg/db"
	{{- end}}
	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/onexstack/onexstack/pkg/server"
	"github.com/onexstack/onexstack/pkg/store/registry"
	"gorm.io/gorm"

	"{{.M.ModuleName}}/internal/{{.MQ.Name}}/handler"
    {{- if .MQ.WithPreloader}}
	"{{.M.ModuleName}}/internal/{{.MQ.Name}}/pkg/asyncstore"
	{{- end}}
	{{- range .MQ.Clients }}
	"{{$.M.ModuleName}}/internal/{{$.MQ.Name}}/pkg/clientset/typed/{{. | lowerkind}}"
	{{- end}}
	{{- if .MQ.WithOTel}}
    "{{.M.ModuleName}}/internal/{{.MQ.Name}}/pkg/metrics"
    {{- end}}
)

const serviceName = "{{.MQ.BinaryName}}"

// Dependencies collects all components that need initialization but are not directly used
// by the main server struct during runtime (e.g., sidecar processes, cache warmers).
type Dependencies struct{}

// Config contains application-related configurations.
type Config struct {
	ObservabilityOptions *genericoptions.ObservabilityOptions
	{{- if eq .MQ.MQFramework "kafka"}}
	KafkaOptions         *genericoptions.KafkaOptions
	{{- end}}
	{{- if eq .MQ.StorageType "mariadb" }}
	MySQLOptions      *genericoptions.MySQLOptions
	{{- end}}
	{{- if eq .MQ.StorageType "postgresql" }}
	PostgreSQLOptions *genericoptions.PostgreSQLOptions
	{{- end}}
	{{- if eq .MQ.StorageType "sqlite" }}
	SQLiteOptions *genericoptions.SQLiteOptions
	{{- end}}
	{{- range .MQ.Clients }}
	{{. | kind}}Options *genericoptions.RestyOptions	
	{{- end}}
}

// Server represents the web server and its background workers.
type Server struct {
    cfg         *ServerConfig
    srv         server.Server
}

// ServerConfig contains the core dependencies and configurations of the server.
type ServerConfig struct {
	*Config
    Dependencies *Dependencies
    Handler      *handler.Handler
}

// New creates and returns a new Server instance.
func (cfg *Config) New(ctx context.Context) (*Server, error) {
    // Create the core server instance using dependency injection.
    // This relies on the wire-generated NewServer function.
    s, err := NewServer(ctx, cfg)
    if err != nil {
        return nil, err
    }

    return s.Prepare(ctx)
}

// Prepare performs post-initialization tasks such as registering subscribers.
func (s *Server) Prepare(ctx context.Context) (*Server, error) {
	{{- if .MQ.WithOTel}}
	metrics.Init(serviceName)
	{{- end}}
    return s, nil
}

// Run starts the server and listens for termination signals.
// It gracefully shuts down the server upon receiving a termination signal from the context.
func (s *Server) Run(ctx context.Context) error {
	// Start the HTTP/gRPC server in a background goroutine.
	go s.srv.RunOrDie(ctx)

	// Block until the context is canceled (e.g., via SIGINT/SIGTERM).
	<-ctx.Done()

	slog.Info("shutting down server...")

    // Create a new context with a timeout to ensure graceful shutdown doesn't hang indefinitely.
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Trigger graceful shutdown for all components.
	s.srv.GracefulStop(shutdownCtx)

	slog.Info("server exited successfully")

	return nil
}

// NewDB creates and returns a *gorm.DB instance for database operations.
func (cfg *Config) NewDB() (*gorm.DB, error) {
	slog.Info("initializing database connection", "type", "{{.MQ.StorageType}}")

	{{- if eq .MQ.StorageType "mariadb" }}
	dbInstance, err := cfg.MySQLOptions.NewDB()
	{{- end}}
	{{- if eq .MQ.StorageType "postgresql" }}
	dbInstance, err := cfg.PostgreSQLOptions.NewDB()
	{{- end}}
	{{- if eq .MQ.StorageType "sqlite" }}
	dbInstance, err := cfg.SQLiteOptions.NewDB()
	{{- end}}
	{{- if eq .MQ.StorageType "memory" }}
	// TODO: Retrieve the database path from configuration instead of hardcoding.
	dbInstance, err := db.NewInMemorySQLite("/tmp/{{.MQ.BinaryName}}.db")
	{{- end}}
	if err != nil {
		slog.Error("failed to create database connection", "error", err)
		return nil, err
	}

	// Automatically migrate database schema
	if err := registry.Migrate(dbInstance); err != nil {
		slog.Error("failed to migrate database schema", "error", err)
		return nil, err
	}

	return dbInstance, nil
}

// ProvideDB provides a database instance based on the configuration.
func ProvideDB(cfg *Config) (*gorm.DB, error) {
	return cfg.NewDB()
}

{{- range .MQ.Clients }}
// Provide{{. | kind}}Client creates and returns a {{. | lowerkind}} client instance using the provided configuration.
func Provide{{. | kind}}Client(cfg *Config) {{. | lowerkind}}.Interface {
    return {{. | lowerkind}}.NewForConfig(cfg.{{. | kind}}Options)
}
{{- end}}

{{- if .MQ.WithPreloader}}
// ProvideAStore creates and returns an asynchronous store factory.
func ProvideAStore(ctx context.Context) asyncstore.Factory {
	return asyncstore.NewStore(ctx, 30*time.Minute)
}
{{- end}}

// NewDependencies initializes all components that need to be started but are not directly stored.
// This is typically used for side-effects or warming up caches.
func NewDependencies(ctx context.Context{{- if .MQ.WithPreloader }}, _ asyncstore.Factory{{- end -}}) *Dependencies {
	{{- if .MQ.WithPreloader}}
	// Simulate cache warmup or check.
	fakeItem, _ := asyncstore.S.Fake().Get("fixed-item-001")
	slog.DebugContext(ctx, "successfully retrieved fake cache data", "data", fakeItem.String())
	{{- end}}

	return &Dependencies{}
}

// NewMQServer creates and returns a new web server instance using the provided server configuration.
func NewMQServer(serverConfig *ServerConfig) (server.Server, error) {
    return serverConfig.NewMQServer()
}
