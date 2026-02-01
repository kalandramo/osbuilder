{{- $D := or .Web .MQ .Job -}}
package model

import (
    "time"
)

// TableName{{$D.R.SingularName}} defines the physical table name for the {{$D.R.GORMModel}} model.
const TableName{{$D.R.SingularName}} = "{{$D.BinaryName | extractProjectPrefix}}_{{$D.R.SingularLower}}"

// {{$D.R.GORMModel}} represents the data model for the {{$D.R.SingularLower}} resource.
// It maps to the "{{$D.BinaryName | extractProjectPrefix}}_{{$D.R.SingularLower}}" table in the database.
type {{$D.R.GORMModel}} struct {
    // ID is the primary key of the record.
    ID int64 `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`

    // {{$D.R.SingularName}}ID is the unique identifier for the resource.
    {{$D.R.SingularName}}ID string `gorm:"column:{{$D.R.SingularLower}}_id;not null;comment:Unique resource ID" json:"{{$D.R.SingularLower}}_id"`

    // CreatedAt is the timestamp when the resource was created.
    CreatedAt time.Time `gorm:"column:created_at;not null;default:current_timestamp;comment:Creation timestamp" json:"createdAt"`

    // UpdatedAt is the timestamp when the resource was last modified.
    UpdatedAt time.Time `gorm:"column:updated_at;not null;default:current_timestamp;comment:Last modification timestamp" json:"updatedAt"`
}

// TableName returns the physical table name for the {{$D.R.GORMModel}} model.
func (*{{$D.R.GORMModel}}) TableName() string {
    return TableName{{$D.R.SingularName}}
}
