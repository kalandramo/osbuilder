{{- $D := or .Web .MQ .Job -}}
package {{$D.R.SingularLower}}

import (
	"context"
	"log/slog"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/onexstack/onexstack/pkg/store/where"
	"github.com/onexstack/onexstack/pkg/watch"
	"github.com/onexstack/onexstack/pkg/watch/registry"

    "{{.M.ModuleName}}/internal/{{$D.Name}}/model"                    
    "{{.M.ModuleName}}/internal/{{$D.Name}}/store"                       
    "{{.M.ModuleName}}/internal/{{$D.Name}}/pkg/metrics"                     
    "{{.M.ModuleName}}/internal/{{$D.Name}}/watcher"                     
    "{{.M.ModuleName}}/internal/pkg/contextx"                                  
)

const (
	// DefaultWatchInterval is the default interval for checking jobs.
	DefaultWatchInterval = "@every 1s"
	{{$D.R.SingularName}}Watcher = "{{$D.R.SingularName}}"
)

// Watcher implements a simplified {{$D.R.SingularLower}} monitoring system.
type Watcher struct {
	watch.Watcher

	store store.IStore
}

// Ensure Watcher implements all required interfaces
var _ registry.Watcher = (*Watcher)(nil)

// Run executes the watcher with concurrency control, and rate limiting.
func (w *Watcher) Run() {
	_, {{$D.R.PluralLower}}, err := w.store.{{$D.R.SingularName}}().List(context.Background(), where.NewWhere())
	if err != nil {
		slog.Error("Failed to list {{$D.R.PluralLower}}", "error", err)
		return
	}

	wp := workerpool.New(w.PerConcurrency)
	for _, {{$D.R.SingularLower}} := range {{$D.R.PluralLower}} {
		ctx := contextx.WithLogger(context.Background(), slog.With("{{$D.R.SingularLower}}_id", {{$D.R.SingularLower}}.{{$D.R.SingularName}}ID))
		wp.Submit(func() { _ = w.processJob(ctx, {{$D.R.SingularLower}}) })
	}

	wp.StopWait()
}

// processJob processes a single {{$D.R.SingularLower}} with error handling and rate limiting.
func (w *Watcher) processJob(ctx context.Context, {{$D.R.SingularLowerFirst}}M *model.{{$D.R.GORMModel}}) (err error) {
	w.Limiter.Take()

	metrics.M.RecordJobExecution(ctx, "{{$D.R.SingularLower}}", {{$D.R.SingularLowerFirst}}M.{{$D.R.SingularName}}ID)
	defer func() {
		if err != nil {
			metrics.M.RecordJobFailure(ctx, "{{$D.R.SingularLower}}", {{$D.R.SingularLowerFirst}}M.{{$D.R.SingularName}}ID, err.Error())
		}
	}()

	contextx.L(ctx).InfoContext(ctx, "Started to process {{$D.R.SingularLower}} job")

	// Simulate {{$D.R.SingularLower}} processing (replace with actual business logic)
	// Add your actual {{$D.R.SingularLower}} processing logic here
	time.Sleep(100 * time.Millisecond)

	contextx.L(ctx).InfoContext(ctx, "Ended to process {{$D.R.SingularLower}} job")

	return nil
}

// Spec returns the cron job specification for scheduling.
func (w *Watcher) Spec() string {
	return DefaultWatchInterval
}

// SetAggregateConfig initializes the watcher for later execution.
func (w *Watcher) SetAggregateConfig(config *watcher.AggregateConfig) {
	w.store = config.Store
}

func init() {
	registry.Register({{$D.R.SingularName}}Watcher, &Watcher{})
}
