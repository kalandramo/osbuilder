package llmtrain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/looplab/fsm"
	"k8s.io/utils/ptr"

	"{{.M.ModuleName}}/internal/{{.Job.Name}}/model"
	"{{.M.ModuleName}}/internal/{{.Job.Name}}/pkg/clientset/typed/fakeminio"
	"{{.M.ModuleName}}/internal/{{.Job.Name}}/pkg/clientset/typed/train"
	"{{.M.ModuleName}}/internal/{{.Job.Name}}/pkg/path"
	"{{.M.ModuleName}}/internal/pkg/contextx"
	known "{{.M.ModuleName}}/internal/pkg/known/job"
	jobconditionsutil "{{.M.ModuleName}}/internal/pkg/util/jobconditions"
    {{.Job.APIImportPath}}
)

// Download fetches feedback data from the Voice of Customer (VOC) system and persists it to object storage.
func (sm *StateMachine) Download(ctx context.Context, event *fsm.Event) error {
	SetDefaultJobParams(sm.Job)

	if ShouldSkipOnIdempotency(sm.Job, event.FSM.Current()) {
		return nil
	}

	time.Sleep(2 * time.Second)

	// Initialize job results structure if not present.
	if sm.Job.Results == nil || sm.Job.Results.Train == nil {
		sm.Job.Results = &model.JobResults{
			Train: &v1.TrainResults{},
		}
	}

	data, err := sm.Watcher.Minio.Read(ctx, fakeminio.FakeObjectName)
	if err != nil {
		return err
	}

	dataPath := path.Job.Path(sm.Job.JobID, path.JobDataName)
	if err := sm.Watcher.Minio.Write(ctx, dataPath, data); err != nil {
		return err
	}

	sm.Job.Results.Train.DataPath = &dataPath
	sm.Job.Conditions = jobconditionsutil.Set(sm.Job.Conditions, jobconditionsutil.TrueCondition(event.FSM.Current()))

	return nil
}

// Embedding processes the daily estimation data to generate vector embeddings.
func (sm *StateMachine) Embedding(ctx context.Context, event *fsm.Event) error {
	if ShouldSkipOnIdempotency(sm.Job, event.FSM.Current()) {
		return nil
	}

	results := sm.Job.Results.Train
	if results.DataPath == nil {
		return fmt.Errorf("data path is nil")
	}

	docs, err := sm.Watcher.Minio.Read(ctx, *results.DataPath)
	if err != nil {
		return err
	}

	embs, err := sm.Watcher.embedder.EmbedDocuments(ctx, docs)
	if err != nil {
		// Log error with structured attributes instead of string formatting.
		contextx.L(ctx).ErrorContext(ctx, "Failed to embed documents", "error", err)
		return err
	}

	// Convert embeddings to JSON strings for better interoperability than raw "%v".
	embeddingsStr := make([]string, 0, len(embs))
	for _, emb := range embs {
		// Assuming emb is compatible with json.Marshal (e.g., []float32).
		bytes, err := json.Marshal(emb)
		if err != nil {
			contextx.L(ctx).ErrorContext(ctx, "Failed to marshal embedding", "error", err)
			return err
		}
		embeddingsStr = append(embeddingsStr, string(bytes))
	}

	embeddedDataPath := path.Job.Path(sm.Job.JobID, path.JobEmbeddedDataName)
	if err := sm.Watcher.Minio.Write(ctx, embeddedDataPath, embeddingsStr); err != nil {
		return err
	}

	results.EmbeddedDataPath = &embeddedDataPath
	// Reset TaskID to ensure subsequent training steps pick up the new data.
	results.TaskID = nil

	// Clean up previous state conditions if necessary.
	jobconditionsutil.Delete(sm.Job.Conditions, known.LLMTrainTrained)
	sm.Job.Conditions = jobconditionsutil.Set(sm.Job.Conditions, jobconditionsutil.TrueCondition(event.FSM.Current()))

	return nil
}

// Train submits the embedded data to the training platform and monitors the task status.
func (sm *StateMachine) Train(ctx context.Context, event *fsm.Event) error {
	if ShouldSkipOnIdempotency(sm.Job, event.FSM.Current()) {
		return nil
	}

	// Apply rate limiting.
	_ = sm.Watcher.Limiter.Take()

	results := sm.Job.Results.Train
	if results.EmbeddedDataPath == nil {
		return fmt.Errorf("embedded data path is nil")
	}

	// Create the training task if not already created.
	if results.TaskID == nil {
		resultPath := path.Job.Path(sm.Job.JobID, path.JobResultName)
		taskID, err := sm.Watcher.train.CreateTask(ctx, *results.EmbeddedDataPath, resultPath)
		if err != nil {
			return err
		}
		results.TaskID = &taskID
		results.ResultPath = &resultPath
	}

	status, err := sm.Watcher.train.GetTaskStatus(ctx, *results.TaskID)
	if err != nil {
		contextx.L(ctx).ErrorContext(ctx, "Failed to get task status", "task_id", *results.TaskID, "error", err)
		return err
	}

	if status != train.StatusCompleted {
		contextx.L(ctx).InfoContext(ctx, "Train task is still running", "task_id", *results.TaskID, "current_status", status)
		// Keep the FSM in the current state to retry later.
		event.FSM.SetState(event.Event)
		return nil
	}

	sm.Job.Conditions = jobconditionsutil.Set(sm.Job.Conditions, jobconditionsutil.TrueCondition(event.FSM.Current()))
	return nil
}

// EnterState handles state transition logic, including status updates, timestamps, and error handling.
func (sm *StateMachine) EnterState(ctx context.Context, event *fsm.Event) error {
	sm.Job.Status = event.FSM.Current()

	// Update lifecycle timestamps.
	now := time.Now()
	switch sm.Job.Status {
	case known.LLMTrainDownloading:
		sm.Job.StartedAt = ptr.To(now)
	case known.LLMTrainSucceeded:
		sm.Job.EndedAt = ptr.To(now)
	}

	// Handle job failures or timeouts.
	if event.Err != nil || isJobTimeout(sm.Job) {
		sm.Job.Status = known.LLMTrainFailed
		sm.Job.EndedAt = ptr.To(now)

		var cond *v1.JobCondition
		if isJobTimeout(sm.Job) {
			contextx.L(ctx).InfoContext(ctx, "LLM train task timed out")
			cond = jobconditionsutil.FalseCondition(event.FSM.Current(), "LLM train task exceeded timeout seconds")
		} else {
			// Note: event.Err is internal error info, handled by the caller or logged elsewhere.
			cond = jobconditionsutil.FalseCondition(event.FSM.Current(), event.Err.Error())
		}

		sm.Job.Conditions = jobconditionsutil.Set(sm.Job.Conditions, cond)
	}

	// Persist the updated job state.
	if err := sm.Watcher.Store.Job().Update(ctx, sm.Job); err != nil {
		return err
	}

	return nil
}
