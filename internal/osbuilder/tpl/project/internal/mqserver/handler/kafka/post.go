package handler

import (
	"context"
	"fmt"
    "log/slog"

	{{- if eq .MQ.MQFramework "kafka"}}
	"github.com/twmb/franz-go/pkg/kgo"
	{{- end}}
	"google.golang.org/protobuf/encoding/protojson"
	{{- if .MQ.WithOTel}}
    "go.opentelemetry.io/otel"

    "{{.M.ModuleName}}/internal/{{.MQ.Name}}/pkg/metrics"
    {{- end}}
	{{- if eq .MQ.MQFramework "kafka"}}
	"{{.M.ModuleName}}/internal/pkg/kafka"
    {{- end}}
	{{.MQ.APIImportPath}}
)

// On{{.MQ.R.SingularName}}Event processes {{.MQ.R.SingularLower}} events from kafka.
func (h *Handler) On{{.MQ.R.SingularName}}Event(ctx context.Context, r *kgo.Record) error {
	envelope := &v1.{{.MQ.R.SingularName}}EventEnvelope{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(r.Value, envelope); err != nil {
		return fmt.Errorf("unmarshal {{.MQ.R.SingularLower}} event failed: %w", err)
	}

	slog.InfoContext(ctx, "processing {{.MQ.R.SingularLower}} event", "event_id", envelope.EventID, "type", envelope.Type)

	switch envelope.Type {
	case v1.{{.MQ.R.SingularName}}EventType_{{.MQ.R.SingularName | toupper}}_EVENT_TYPE_CREATED:
		return h.Create{{.MQ.R.SingularName}}(ctx, envelope.GetCreated())
	case v1.{{.MQ.R.SingularName}}EventType_{{.MQ.R.SingularName | toupper}}_EVENT_TYPE_UPDATED:
		return h.Update{{.MQ.R.SingularName}}(ctx, envelope.GetUpdated())
	case v1.{{.MQ.R.SingularName}}EventType_{{.MQ.R.SingularName | toupper}}_EVENT_TYPE_DELETED:
		return h.Delete{{.MQ.R.SingularName}}(ctx, envelope.GetDeleted())
	case v1.{{.MQ.R.SingularName}}EventType_{{.MQ.R.SingularName | toupper}}_EVENT_TYPE_COLLECTION_DELETED:
		return h.Delete{{.MQ.R.PluralName}}(ctx, envelope.GetCollectionDeleted())
	default:
		slog.WarnContext(ctx, "ignored unknown {{.MQ.R.SingularLower}} event type", "type", envelope.Type)
		return nil
	}
}

// Create{{.MQ.R.SingularName}} handles the kafka message to create a new {{.MQ.R.SingularLower}}.
func (h *Handler) Create{{.MQ.R.SingularName}}(ctx context.Context, rq *v1.Create{{.MQ.R.SingularName}}Request) error {
	{{- if .MQ.WithOTel}}
	ctx, span := otel.Tracer("handler").Start(ctx, "Handler.Create{{.MQ.R.SingularName}}")
	defer span.End()

	metrics.M.RecordResourceCreate(ctx, "{{.MQ.R.SingularLower}}")
	{{- end}}

	slog.InfoContext(ctx, "processing {{.MQ.R.SingularLower}} creation request")

	if err := h.val.ValidateCreate{{.MQ.R.SingularName}}Request(ctx, rq); err != nil {
		return err
	}

	_, err := h.biz.{{.MQ.R.BusinessFactoryName}}().Create(ctx, rq)
	return err
}

// Update{{.MQ.R.SingularName}} handles the kafka message to update an existing {{.MQ.R.SingularLower}}'s details.
func (h *Handler) Update{{.MQ.R.SingularName}}(ctx context.Context, rq *v1.Update{{.MQ.R.SingularName}}Request) error {
	if err := h.val.ValidateUpdate{{.MQ.R.SingularName}}Request(ctx, rq); err != nil {
		return err
	}

	_, err := h.biz.{{.MQ.R.BusinessFactoryName}}().Update(ctx, rq)
	return err
}

// Delete{{.MQ.R.SingularName}} handles the kafka message to delete a single {{.MQ.R.SingularLower}}.
func (h *Handler) Delete{{.MQ.R.SingularName}}(ctx context.Context, rq *v1.Delete{{.MQ.R.SingularName}}Request) error {
	if err := h.val.ValidateDelete{{.MQ.R.SingularName}}Request(ctx, rq); err != nil {
		return err
	}

	_, err := h.biz.{{.MQ.R.BusinessFactoryName}}().Delete(ctx, rq)
	return err
}

// Delete{{.MQ.R.PluralName}} handles the kafka message to delete a collection of {{.MQ.R.PluralName}}.
func (h *Handler) Delete{{.MQ.R.PluralName}}(ctx context.Context, rq *v1.Delete{{.MQ.R.PluralName}}Request) error {
	if err := h.val.ValidateDelete{{.MQ.R.PluralName}}Request(ctx, rq); err != nil {
		return err
	}

	_, err := h.biz.{{.MQ.R.BusinessFactoryName}}().DeleteCollection(ctx, rq)
	return err
}

func init() {
    Register(func(engine *kafka.Engine, h *Handler) {
        engine.Register("{{.MQ.BinaryName | toDot}}.{{.MQ.R.SingularLower}}.events.v1", h.On{{.MQ.R.SingularName}}Event)
    })
}
