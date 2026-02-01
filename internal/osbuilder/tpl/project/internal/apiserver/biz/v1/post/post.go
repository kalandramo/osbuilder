{{- $D := or .Web .MQ .Job -}}
package {{$D.R.SingularLower}}

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/onexstack/onexstack/pkg/core"
	"github.com/onexstack/onexstack/pkg/store/where"
	{{- if $D.WithOTel}}
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    oteltrace "go.opentelemetry.io/otel/trace"
    {{- end}}
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"{{.M.ModuleName}}/internal/{{$D.Name}}/model"
	"{{.M.ModuleName}}/internal/{{$D.Name}}/pkg/conversion"
	"{{.M.ModuleName}}/internal/pkg/known"
	"{{.M.ModuleName}}/internal/pkg/errno"
	"{{.M.ModuleName}}/internal/{{$D.Name}}/store"
	// "{{.M.ModuleName}}/internal/pkg/contextx"
	{{$D.APIImportPath}}
    {{- if $D.Clients }}
    "{{.M.ModuleName}}/internal/{{$D.Name}}/pkg/clientset"
    {{- end}}
)

// {{$D.R.SingularName}}Biz defines the interface for handling {{$D.R.SingularLower}}-related business logic.
type {{$D.R.SingularName}}Biz interface {
	// Create creates a new {{$D.R.SingularLower}} based on the provided request parameters.
	Create(ctx context.Context, rq *{{.M.APIAlias}}.Create{{$D.R.SingularName}}Request) (*{{.M.APIAlias}}.Create{{$D.R.SingularName}}Response, error)

	// Update updates an existing {{$D.R.SingularLower}} based on the provided request parameters.
	Update(ctx context.Context, rq *{{.M.APIAlias}}.Update{{$D.R.SingularName}}Request) (*{{.M.APIAlias}}.Update{{$D.R.SingularName}}Response, error)

	// Delete remove one {{$D.R.SingularLower}} based on the provided request parameters.
	Delete(ctx context.Context, rq *{{.M.APIAlias}}.Delete{{$D.R.SingularName}}Request) (*{{.M.APIAlias}}.Delete{{$D.R.SingularName}}Response, error)

	// DeleteCollection deletes a collection of {{$D.R.PluralLower}} that match the specified criteria or identifiers.
	DeleteCollection(ctx context.Context, rq *v1.Delete{{$D.R.PluralName}}Request) (*v1.Delete{{$D.R.PluralName}}Response, error)

	// Get retrieves the details of a specific {{$D.R.SingularLower}} based on the provided request parameters.
	Get(ctx context.Context, rq *{{.M.APIAlias}}.Get{{$D.R.SingularName}}Request) (*{{.M.APIAlias}}.Get{{$D.R.SingularName}}Response, error)

	// List retrieves a list of {{$D.R.PluralLower}} and their total count based on the provided request parameters.
	List(ctx context.Context, rq *{{.M.APIAlias}}.List{{$D.R.SingularName}}Request) (*{{.M.APIAlias}}.List{{$D.R.SingularName}}Response, error)

	// {{$D.R.SingularName}}Expansion defines additional methods for extended {{$D.R.SingularLower}} operations, if needed.
	{{$D.R.SingularName}}Expansion
}

// {{$D.R.SingularName}}Expansion defines custom methods for extended {{$D.R.SingularLower}} business operations.
type {{$D.R.SingularName}}Expansion interface{}

// {{$D.R.SingularLowerFirst}}Biz implements the {{$D.R.SingularName}}Biz interface.
type {{$D.R.SingularLowerFirst}}Biz struct {
	store store.IStore
	{{- if $D.Clients }}
	clientset clientset.Interface
	{{- end}}
}

// Ensure {{$D.R.SingularLowerFirst}}Biz implements {{$D.R.SingularName}}Biz at compile time.
var _ {{$D.R.SingularName}}Biz = (*{{$D.R.SingularLowerFirst}}Biz)(nil)

// New creates and returns a new instance of {{$D.R.SingularName}}Biz.
func New(store store.IStore{{- if $D.Clients }}, clientset clientset.Interface{{- end -}}) *{{$D.R.SingularLowerFirst}}Biz {
	return &{{$D.R.SingularLowerFirst}}Biz{store: store{{- if $D.Clients}}, clientset: clientset{{- end -}}}
}

// Create implements the Create method of the {{$D.R.SingularName}}Biz.
func (b *{{$D.R.SingularLowerFirst}}Biz) Create(ctx context.Context, rq *{{.M.APIAlias}}.Create{{$D.R.REST.SingularName}}Request) (*{{.M.APIAlias}}.Create{{$D.R.SingularName}}Response, error) {
	{{- if $D.WithOTel}}
    ctx, span := otel.Tracer("biz").Start(ctx, "{{$D.R.SingularLowerFirst}}Biz.Create")
    defer span.End()

    // Follow the component.operation.phase pattern
    span.AddEvent("{{$D.R.SingularLower}}.creation.started")
    {{- end}}

	var {{$D.R.SingularLowerFirst}}M model.{{$D.R.GORMModel}}
	_ = core.Copy(&{{$D.R.SingularLowerFirst}}M, rq)
	// TODO: Retrieve the UserID from the custom context and assign it as needed.
	// {{$D.R.SingularLowerFirst}}M.UserID = contextx.UserID(ctx)
                                                                                
	slog.InfoContext(ctx, "creating {{$D.R.SingularLower}} in database")

	if err := b.store.{{$D.R.SingularName}}().Create(ctx, &{{$D.R.SingularLowerFirst}}M); err != nil {
    	{{- if $D.WithOTel}}
		core.RecordSpanError(ctx, span, err)
    	{{- end}}
		slog.ErrorContext(ctx, "failed to create {{$D.R.SingularLower}}", "error", err)
		return nil, errno.Err{{$D.R.SingularName}}CreateFailed.WithMessage(err.Error())
	}

	{{- if $D.WithOTel}}
	span.AddEvent("{{$D.R.SingularLower}}.creation.completed", oteltrace.WithAttributes(attribute.String("{{$D.R.SingularLower}}_id", {{$D.R.SingularLowerFirst}}M.{{$D.R.SingularName}}ID)))
    {{- end}}
	return &{{.M.APIAlias}}.Create{{$D.R.SingularName}}Response{ {{$D.R.SingularName}}ID: {{$D.R.SingularLowerFirst}}M.{{$D.R.SingularName}}ID}, nil
}

// Update implements the Update method of the {{$D.R.SingularName}}Biz.
func (b *{{$D.R.SingularLowerFirst}}Biz) Update(ctx context.Context, rq *{{.M.APIAlias}}.Update{{$D.R.SingularName}}Request) (*{{.M.APIAlias}}.Update{{$D.R.SingularName}}Response, error) {
	whr := where.F("{{$D.R.SingularLower}}_id", rq.{{$D.R.SingularName}}ID)
	{{$D.R.SingularLowerFirst}}M, err := b.store.{{$D.R.SingularName}}().Get(ctx, whr)
	if err != nil {
		return nil, errno.Err{{$D.R.SingularName}}UpdateFailed.WithMessage(err.Error())
	}

    // TODO: Apply updates to {{$D.R.SingularLowerFirst}}M from rq.
    // Example: {{$D.R.SingularLowerFirst}}M.Status = rq.Status

	if err := b.store.{{$D.R.SingularName}}().Update(ctx, {{$D.R.SingularLowerFirst}}M); err != nil {
		return nil, errno.Err{{$D.R.SingularName}}UpdateFailed.WithMessage(err.Error())
	}

	return &{{.M.APIAlias}}.Update{{$D.R.SingularName}}Response{}, nil
}

// Delete implements the Delete method of the {{$D.R.SingularName}}Biz.
func (b *{{$D.R.SingularLowerFirst}}Biz) Delete(ctx context.Context, rq *{{.M.APIAlias}}.Delete{{$D.R.SingularName}}Request) (*{{.M.APIAlias}}.Delete{{$D.R.SingularName}}Response, error) {
	whr := where.F("{{$D.R.SingularLower}}_id", rq.{{$D.R.SingularName}}ID)
	if err := b.store.{{$D.R.SingularName}}().Delete(ctx, whr); err != nil {
		return nil, errno.Err{{$D.R.SingularName}}DeleteFailed.WithMessage(err.Error())
	}

	return &{{.M.APIAlias}}.Delete{{$D.R.SingularName}}Response{}, nil
}

// DeleteCollection implements the DeleteCollection method of the {{$D.R.SingularName}}Biz.
func (b *{{$D.R.SingularLowerFirst}}Biz) DeleteCollection(ctx context.Context, rq *v1.Delete{{$D.R.PluralName}}Request) (*v1.Delete{{$D.R.PluralName}}Response, error) {
    whr := where.F("{{$D.R.SingularLower}}_id", rq.{{$D.R.SingularName}}IDs)
    if err := b.store.{{$D.R.SingularName}}().Delete(ctx, whr); err != nil {
        return nil, errno.Err{{$D.R.SingularName}}DeleteFailed.WithMessage(err.Error())
    }

    return &v1.Delete{{$D.R.PluralName}}Response{}, nil
}

// Get implements the Get method of the {{$D.R.SingularName}}Biz.
func (b *{{$D.R.SingularLowerFirst}}Biz) Get(ctx context.Context, rq *{{.M.APIAlias}}.Get{{$D.R.SingularName}}Request) (*{{.M.APIAlias}}.Get{{$D.R.SingularName}}Response, error) {
	{{- if $D.WithOTel}}
    ctx, span := otel.Tracer("biz").Start(ctx, "{{$D.R.SingularLowerFirst}}Biz.Get")
    defer span.End()

	span.SetAttributes(attribute.String("{{$D.R.SingularLowerFirst}}_id", rq.{{$D.R.SingularName}}ID))
    {{- end}}

	slog.InfoContext(ctx, "retrieving job from database", "job_id", rq.{{$D.R.SingularName}}ID)

	whr := where.F("{{$D.R.SingularLower}}_id", rq.{{$D.R.SingularName}}ID)
	{{$D.R.SingularLowerFirst}}M, err := b.store.{{$D.R.SingularName}}().Get(ctx, whr)
	if err != nil {
		{{- if $D.WithOTel}}
		core.RecordSpanError(ctx, span, err)
    	{{- end}}
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errno.Err{{$D.R.SingularName}}NotFound 
        }                       
		slog.ErrorContext(ctx, "failed to retrive {{$D.R.SingularLower}}", "error", err, "{{$D.R.SingularLower}}_id", rq.{{$D.R.SingularName}}ID)
		return nil, errno.Err{{$D.R.SingularName}}GetFailed.WithMessage(err.Error())
	}

	return &{{.M.APIAlias}}.Get{{$D.R.SingularName}}Response{ {{$D.R.SingularName}}: conversion.{{$D.R.MapModelToAPIFunc}}({{$D.R.SingularLowerFirst}}M)}, nil
}

// List implements the List method of the {{$D.R.SingularName}}Biz.
func (b *{{$D.R.SingularLowerFirst}}Biz) List(ctx context.Context, rq *{{.M.APIAlias}}.List{{$D.R.SingularName}}Request) (*{{.M.APIAlias}}.List{{$D.R.SingularName}}Response, error) {
	whr := where.P(int(rq.Offset), int(rq.Limit))
	count, {{$D.R.SingularLowerFirst}}List, err := b.store.{{$D.R.SingularName}}().List(ctx, whr)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list {{$D.R.PluralLower}}", "error", err)
		return nil, errno.Err{{$D.R.SingularName}}ListFailed.WithMessage(err.Error())
	}

	// Concurrent processing for list items conversion/enrichment.
	var m sync.Map
	eg, ctx := errgroup.WithContext(ctx)

	// Set the maximum concurrency limit using the constant MaxConcurrency
	eg.SetLimit(known.MaxErrGroupConcurrency)

	// Use goroutines to improve API performance
	for _, {{$D.R.SingularLowerFirst}} := range {{$D.R.SingularLowerFirst}}List {
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				converted := conversion.{{$D.R.MapModelToAPIFunc}}({{$D.R.SingularLowerFirst}})
				// TODO: Add complex enrichment logic here if needed.
				m.Store({{$D.R.SingularLowerFirst}}.ID, converted)

				return nil
			}
		})
	}

	if err := eg.Wait(); err != nil {
		slog.ErrorContext(ctx, "error during concurrent {{$D.R.SingularLower}} processing", "error", err)
		return nil, errno.Err{{$D.R.SingularName}}ListFailed.WithMessage(err.Error())
	}

	// Reassemble the result in the correct order.
	{{$D.R.PluralLowerFirst}} := make([]*{{.M.APIAlias}}.{{$D.R.SingularName}}, 0, len({{$D.R.SingularLowerFirst}}List))
	for _, item := range {{$D.R.SingularLowerFirst}}List {
        if val, ok := m.Load(item.ID); ok {
            {{$D.R.PluralLowerFirst}} = append({{$D.R.PluralLowerFirst}}, val.(*{{.M.APIAlias}}.{{$D.R.SingularName}}))
        }
	}

	return &{{.M.APIAlias}}.List{{$D.R.SingularName}}Response{Total: count, {{$D.R.PluralName}}: {{$D.R.PluralLowerFirst}}}, nil
}
