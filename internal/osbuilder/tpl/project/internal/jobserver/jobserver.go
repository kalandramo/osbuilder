package {{.Job.Name}}

import (
	"context"

	"github.com/onexstack/onexstack/pkg/server"
    "github.com/onexstack/onexstack/pkg/watch"
    sloglogger "github.com/onexstack/onexstack/pkg/watch/logger/slog"

    {{- if not .Job.DisableCronJob }}
    "{{.M.ModuleName}}/internal/{{.Job.Name}}/pkg/clientset/typed/fakeminio"
    {{- end}}
    "{{.M.ModuleName}}/internal/{{.Job.Name}}/store"
    "{{.M.ModuleName}}/internal/{{.Job.Name}}/watcher"
    _ "{{.M.ModuleName}}/internal/{{.Job.Name}}/watcher/all"
)

// jobServer implements the server.Server interface using the kafka framework.
type jobServer struct {
    watch *watch.Watch
}

// Ensure *jobServer implements the server.Server interface.
var _ server.Server = (*jobServer)(nil)

// NewJobServer initializes and returns a new HTTP server based on kafka.
// It returns the server.Server interface to abstract implementation details.
func (c *ServerConfig) NewJobServer() (server.Server, error) {
    {{- if not .Job.DisableCronJob }}
    minio, _ := fakeminio.NewForConfig("fake-bucket")
    {{- end}}
    config := &watcher.AggregateConfig{
        {{- if not .Job.DisableCronJob }}
        Minio: minio,
        {{- end}}
        Store: store.NewStore(c.DB),
    }
 
    initialize := watcher.NewInitializer(config)
    opts := []watch.Option{
        watch.WithInitialize(initialize),
        watch.WithLogger(sloglogger.NewLogger()),
        watch.WithAutoMigrate(true),
        watch.WithHealthzPort(c.WatchOptions.HealthzPort),
        watch.WithMetricsAddr(c.WatchOptions.MetricsAddr),
    }
 
    w, err := watch.NewWatch(c.WatchOptions, c.DB, opts...)
    if err != nil {
        return nil, err
    }

    return &jobServer{watch: w}, nil
}

// RunOrDie starts the Kafka server and panics if startup fails.
func (s *jobServer) RunOrDie(ctx context.Context) {
    s.watch.Start(ctx.Done())
}

// GracefulStop gracefully shuts down the server.
func (s *jobServer) GracefulStop(ctx context.Context) {
    s.watch.Stop()
}
