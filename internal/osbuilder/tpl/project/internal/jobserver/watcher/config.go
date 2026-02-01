// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

// Package watcher provides functions used by all watchers.
package watcher

import (
	{{- if not .Job.DisableCronJob }}
	"{{.M.ModuleName}}/internal/{{.Job.Name}}/pkg/clientset/typed/minio"
	{{- end}}
	"{{.M.ModuleName}}/internal/{{.Job.Name}}/store"
)

// AggregateConfig aggregates the configurations of all watchers and serves as a configuration aggregator.
type AggregateConfig struct {
	{{- if not .Job.DisableCronJob }}
	Minio minio.Interface
	{{- end}}
	Store store.IStore
}
