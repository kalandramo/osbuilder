package all

import (
	_ "{{.M.ModuleName}}/internal/{{.Job.Name}}/watcher/customized/fake"
 	{{- if not .Job.DisableCronJob }}
	_ "{{.M.ModuleName}}/internal/{{.Job.Name}}/watcher/cronjob/cronjob"
	_ "{{.M.ModuleName}}/internal/{{.Job.Name}}/watcher/cronjob/statesync"
	_ "{{.M.ModuleName}}/internal/{{.Job.Name}}/watcher/job/llmtrain"
	{{- end}}
)
