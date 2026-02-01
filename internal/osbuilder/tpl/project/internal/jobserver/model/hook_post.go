{{- $D := or .Web .MQ .Job -}}
package model

import (
	"github.com/onexstack/onexstack/pkg/rid"
	"github.com/onexstack/onexstack/pkg/store/registry"
	"gorm.io/gorm"
)

// AfterCreate generates and updates the {{$D.R.SingularName}}ID after the database record is created.
func (m *{{$D.R.GORMModel}}) AfterCreate(tx *gorm.DB) error {
	// Generate the resource ID based on the auto-increment primary key.
	m.{{$D.R.SingularName}}ID = rid.NewResourceID("{{$D.R.SingularLower}}").New(uint64(m.ID))

	// Update only the {{$D.R.SingularName}}ID column to avoid overhead and side effects of a full Save.
	// UpdateColumn is faster as it doesn't update timestamps or trigger Update hooks.
	return tx.Model(m).UpdateColumn("{{$D.R.SingularLower}}_id", m.{{$D.R.SingularName}}ID).Error
}

func init() {
	registry.Register(&{{$D.R.GORMModel}}{})
}
