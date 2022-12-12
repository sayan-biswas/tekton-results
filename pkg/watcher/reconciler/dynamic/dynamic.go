// Copyright 2020 The Tekton Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dynamic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/jonboulle/clockwork"
	"github.com/tektoncd/cli/pkg/cli"
	tknlog "github.com/tektoncd/cli/pkg/log"
	tknopts "github.com/tektoncd/cli/pkg/options"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/results/pkg/apis/v1alpha2"
	"github.com/tektoncd/results/pkg/watcher/convert"
	"github.com/tektoncd/results/pkg/watcher/logs"
	"github.com/tektoncd/results/pkg/watcher/reconciler"
	"github.com/tektoncd/results/pkg/watcher/reconciler/annotation"
	"github.com/tektoncd/results/pkg/watcher/results"
	pb "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

var (
	clock = clockwork.NewRealClock()
)

// Reconciler implements common reconciler behavior across different Tekton Run
// Object types.
type Reconciler struct {
	resultsClient *results.Client
	logsClient    pb.LogsClient
	objectClient  ObjectClient
	cfg           *reconciler.Config
}

func init() {
	// Disable colorized output from the tkn CLI.
	color.NoColor = true
}

// NewDynamicReconciler creates a new dynamic Reconciler.
func NewDynamicReconciler(rc pb.ResultsClient, oc ObjectClient, cfg *reconciler.Config) *Reconciler {
	return &Reconciler{
		resultsClient: &results.Client{ResultsClient: rc},
		objectClient:  oc,
		cfg:           cfg,
	}
}

// Reconcile handles result/record uploading for the given Run object.
// If enabled, the object may be deleted upon successful result upload.
func (r *Reconciler) Reconcile(ctx context.Context, o results.Object) error {
	logger := logging.FromContext(ctx)

	if o.GetObjectKind().GroupVersionKind().Empty() {
		gvk, err := convert.InferGVK(o)
		if err != nil {
			return err
		}
		o.GetObjectKind().SetGroupVersionKind(gvk)
		logger.Debugf("Post SetGroupVersionKind: %s", o.GetObjectKind().GroupVersionKind().String())
	}

	// Update record.
	result, record, err := r.resultsClient.Put(ctx, o)
	if err != nil {
		logger.Debugw("Error updating Record", zap.Error(err))
		return err
	}

	var logName string

	if !o.GetObjectKind().GroupVersionKind().Empty() && o.GetObjectKind().GroupVersionKind().Kind == "TaskRun" {
		// Create a log record if the object has/supports logs
		// For now this is just TaskRuns.
		logResult, logRecord, err := r.resultsClient.PutLog(ctx, o)
		if err != nil {
			logger.Errorf("error creating TaskRun log Record: %v", err)
			return err
		}
		if logRecord != nil {
			logName = logRecord.GetName()
		}
		needsStream, err := needsLogsStreamed(logRecord)
		if err != nil {
			logger.Errorf("error determining if logs need to be streamed: %v", err)
			return err
		}
		if needsStream {
			logger.Infof("streaming logs for TaskRun %s/%s", o.GetNamespace(), o.GetName())
			err = r.streamLogs(ctx, logResult, logRecord, o)
			if err != nil {
				logger.Errorf("error streaming logs: %v", err)
				return err
			}
			logger.Infof("finished streaming logs for TaskRun %s/%s", o.GetNamespace(), o.GetName())
		}
	}

	logger = logger.With(zap.String("results.tekton.dev/result", result.Name),
		zap.String("results.tekton.dev/record", record.Name))

	if err := r.addResultsAnnotations(logging.WithLogger(ctx, logger), o, result, record); err != nil {
		return err
	}

	return r.deleteUponCompletion(logging.WithLogger(ctx, logger), o)
}

// addResultsAnnotations adds Results annotations to the object in question if
// annotation patching is enabled.
func (r *Reconciler) addResultsAnnotations(ctx context.Context, o results.Object, result *pb.Result, record *pb.Record) error {
	logger := logging.FromContext(ctx)

	objectAnnotations := o.GetAnnotations()
	if r.cfg.GetDisableAnnotationUpdate() {
		logger.Info("Skipping CRD annotation patch: annotation update is disabled")
	} else if result.GetName() == objectAnnotations[annotation.Result] && record.GetName() == objectAnnotations[annotation.Record] {
		logger.Debug("Skipping CRD annotation patch: Result annotations are already set")
	} else {
		// Update object with Result Annotations.
		patch, err := annotation.Add(result.GetName(), record.GetName())
		if err != nil {
			logger.Errorw("Error adding Result annotations", zap.Error(err))
			return err
		}
		if err := r.objectClient.Patch(ctx, o.GetName(), types.MergePatchType, patch, metav1.PatchOptions{}); err != nil {
			logger.Errorw("Error patching object", zap.Error(err))
			return err
		}
	}
	return nil
}

// deleteUponCompletion deletes the object in question when the following
// conditions are met:
// * The resource deletion is enabled in the config (the grace period is greater
// than 0).
// * The object is done, and it isn't owned by other object.
// * The configured grace period has elapsed since the object's completion.
func (r *Reconciler) deleteUponCompletion(ctx context.Context, o results.Object) error {
	logger := logging.FromContext(ctx)

	gracePeriod := r.cfg.GetCompletedResourceGracePeriod()
	logger = logger.With(zap.Duration("results.tekton.dev/gracePeriod", gracePeriod))
	if gracePeriod == 0 {
		logger.Info("Skipping resource deletion: deletion is disabled")
		return nil
	}

	if !isDone(o) {
		logger.Debug("Skipping resource deletion: object is not done yet")
		return nil
	}

	if ownerReferences := o.GetOwnerReferences(); len(ownerReferences) > 0 {
		logger.Debugw("Resource is owned by another object, deferring deletion to parent resource(s)", zap.Any("results.tekton.dev/ownerReferences", ownerReferences))
		return nil
	}

	completionTime, err := getCompletionTime(o)
	if err != nil {
		return err
	}

	// This isn't probable since the object is done, but defensive
	// programming never hurts.
	if completionTime == nil {
		logger.Debug("Object's completion time isn't set yet - requeuing to process later")
		return controller.NewRequeueAfter(gracePeriod)
	}

	if timeSinceCompletion := clock.Since(*completionTime); timeSinceCompletion < gracePeriod {
		requeueAfter := gracePeriod - timeSinceCompletion
		logger.Debugw("Object is not ready for deletion yet - requeuing to process later", zap.Duration("results.tekton.dev/requeueAfter", requeueAfter))
		return controller.NewRequeueAfter(requeueAfter)
	}

	logger.Infow("Deleting object", zap.String("results.tekton.dev/uid", string(o.GetUID())))
	if err := r.objectClient.Delete(ctx, o.GetName(), metav1.DeleteOptions{
		Preconditions: metav1.NewUIDPreconditions(string(o.GetUID())),
	}); err != nil && !errors.IsNotFound(err) {
		logger.Errorw("Error deleting object", zap.Error(err))
		return err
	}

	return nil
}

func (r *Reconciler) streamLogs(ctx context.Context, res *pb.Result, rec *pb.Record, o metav1.Object) error {
	logClient, err := r.logsClient.UpdateLog(ctx)
	if err != nil {
		return fmt.Errorf("failed to create PutLog client: %v", err)
	}
	writer := logs.NewLogWriter(logClient, rec.GetName(), logs.DefaultMaxLogChunkSize)

	tknParams := &cli.TektonParams{}
	tknParams.SetNamespace(o.GetNamespace())
	// KLUGE: tkn reader.Read() will raise an error if a step in the TaskRun failed and there is no
	// Err writer in the Stream object. This will result in some "error" messages being written to
	// the log.

	// TODO: Set TaskrunName or PipelinerunName based on object type
	reader, err := tknlog.NewReader(tknlog.LogTypeTask, &tknopts.LogOptions{
		AllSteps:    true,
		Params:      tknParams,
		TaskrunName: o.GetName(),
		Stream: &cli.Stream{
			Out: writer,
			Err: writer,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create tkn reader: %v", err)
	}
	logChan, errChan, err := reader.Read()
	if err != nil {
		return fmt.Errorf("error reading from tkn reader: %v", err)
	}
	// errChan receives stderr from the TaskRun containers.
	// This will be forwarded as combined output (stdout and stderr)

	// TODO: Set writer type based on the object type
	tknlog.NewWriter(tknlog.LogTypeTask, true).Write(&cli.Stream{
		Out: writer,
		Err: writer,
	}, logChan, errChan)
	return logClient.CloseSend()
}

func isDone(o results.Object) bool {
	return o.GetStatusCondition().GetCondition(apis.ConditionSucceeded).IsTrue()
}

// getCompletionTime returns the completion time of the object (PipelineRun or
// TaskRun) in question.
func getCompletionTime(object results.Object) (*time.Time, error) {
	var completionTime *time.Time

	switch o := object.(type) {

	case *pipelinev1beta1.PipelineRun:
		if o.Status.CompletionTime != nil {
			completionTime = &o.Status.CompletionTime.Time
		}

	case *pipelinev1beta1.TaskRun:
		if o.Status.CompletionTime != nil {
			completionTime = &o.Status.CompletionTime.Time
		}

	default:
		return nil, controller.NewPermanentError(fmt.Errorf("error getting completion time from incoming object: unrecognized type %T", o))
	}
	return completionTime, nil
}

func needsLogsStreamed(rec *pb.Record) (bool, error) {
	if rec.GetData().Type != v1alpha2.TaskRunLogRecordType {
		return false, nil
	}
	trl := &v1alpha2.TaskRunLog{}
	err := json.Unmarshal(rec.GetData().GetValue(), trl)
	if err != nil {
		return false, err
	}
	needsStream := trl.Spec.Type == v1alpha2.FileLogType
	if trl.Status.File != nil {
		needsStream = needsStream && trl.Status.File.Size <= 0
	}
	return needsStream, nil
}
