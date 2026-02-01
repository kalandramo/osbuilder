package model

import (
	"github.com/onexstack/onexstack/pkg/rid"
	"github.com/onexstack/onexstack/pkg/store/registry"
	"gorm.io/gorm"
)

// AfterCreate generates a cronJobID after creating a database record.
func (m *CronJobM) AfterCreate(tx *gorm.DB) error {
	m.CronJobID = rid.NewResourceID("cronjob").New(uint64(m.ID))

	return tx.Save(m).Error
}

func init() {
	registry.Register(&CronJobM{})
}
