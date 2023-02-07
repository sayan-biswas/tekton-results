package pipelinerun

import (
	"context"
	"fmt"

	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"github.com/tektoncd/pipeline/pkg/client/listers/pipeline/v1beta1"
	"github.com/tektoncd/results/pkg/watcher/reconciler"
	"github.com/tektoncd/results/pkg/watcher/reconciler/dynamic"
	pb "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	knativereconciler "knative.dev/pkg/reconciler"
)

type Reconciler struct {
	// Inline LeaderAwareFuncs to support leader election.
	knativereconciler.LeaderAwareFuncs

	resultsClient  pb.ResultsClient
	logsClient     pb.LogsClient
	lister         v1beta1.PipelineRunLister
	pipelineClient versioned.Interface
	cfg            *reconciler.Config
}

// Check that our Reconciler is LeaderAware.
var _ knativereconciler.LeaderAware = (*Reconciler)(nil)

func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx).With(zap.String("results.tekton.dev/kind", "PipelineRun"))

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorf("invalid resource key: %s", key)
		return nil
	}

	if !r.IsLeaderFor(types.NamespacedName{Namespace: namespace, Name: name}) {
		logger.Debug("Skipping PipelineRun key because this instance isn't its leader")
		return controller.NewSkipKey(key)
	}

	logger.Info("Reconciling PipelineRun")

	pr, err := r.lister.PipelineRuns(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Debug("Skipping key: object is no longer available")
			return controller.NewSkipKey(key)
		}
		return fmt.Errorf("error reading PipelineRun from the indexer: %w", err)
	}

	pipelineRunClient := &dynamic.PipelineRunClient{
		PipelineRunInterface: r.pipelineClient.TektonV1beta1().PipelineRuns(namespace),
	}

	dyn := dynamic.NewDynamicReconciler(r.resultsClient, r.logsClient, pipelineRunClient, r.cfg)
	if err := dyn.Reconcile(logging.WithLogger(ctx, logger), pr); err != nil {
		return err
	}

	return nil
}
