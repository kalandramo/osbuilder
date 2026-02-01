{{- $D := or .Web .MQ .Job -}}
// nolint: dupl
package store

import (
	"context"

	storelogger "github.com/onexstack/onexstack/pkg/logger/slog/store"
	genericstore "github.com/onexstack/onexstack/pkg/store"
	"github.com/onexstack/onexstack/pkg/store/where"

	"{{.M.ModuleName}}/internal/{{$D.Name}}/model"
)

// {{$D.R.SingularName}}Store defines the interface for managing {{$D.R.SingularLower}}-related persistent data.
type {{$D.R.SingularName}}Store interface {
	// Create persists a new {{$D.R.SingularLower}} record.
	Create(ctx context.Context, obj *model.{{$D.R.GORMModel}}) error

	// Update modifies an existing {{$D.R.SingularLower}} record.
	Update(ctx context.Context, obj *model.{{$D.R.GORMModel}}) error

	// Delete removes {{$D.R.SingularLower}} records matching the specified criteria.
	Delete(ctx context.Context, opts *where.Options) error

	// Get retrieves a single {{$D.R.SingularLower}} record matching the specified criteria.
	Get(ctx context.Context, opts *where.Options) (*model.{{$D.R.GORMModel}}, error)

	// List retrieves a list of {{$D.R.SingularLower}} records and the total count matching the criteria.
	List(ctx context.Context, opts *where.Options) (int64, []*model.{{$D.R.GORMModel}}, error)

	// {{$D.R.SingularName}}Expansion defines custom methods for the {{$D.R.SingularLower}} store outside the generic CRUD operations.
	{{$D.R.SingularName}}Expansion
}

// {{$D.R.SingularName}}Expansion is an extension interface for {{$D.R.SingularName}}Store.
type {{$D.R.SingularName}}Expansion interface{}

// {{$D.R.SingularLower}}Store implements the {{$D.R.SingularName}}Store interface using a generic store implementation.
type {{$D.R.SingularLower}}Store struct {
	*genericstore.Store[model.{{$D.R.GORMModel}}]
}

// Ensure {{$D.R.SingularLower}}Store implements {{$D.R.SingularName}}Store at compile time.
var _ {{$D.R.SingularName}}Store = (*{{$D.R.SingularLower}}Store)(nil)

// new{{$D.R.SingularName}}Store returns a new instance of {{$D.R.SingularName}}Store.
func new{{$D.R.SingularName}}Store(s *store) *{{$D.R.SingularLower}}Store {
	return &{{$D.R.SingularLower}}Store{
		Store: genericstore.NewStore[model.{{$D.R.GORMModel}}](s, storelogger.NewLogger()),
	}
}
