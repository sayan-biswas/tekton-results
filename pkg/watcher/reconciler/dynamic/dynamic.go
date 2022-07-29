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
	"github.com/jonboulle/clockwork"
	"github.com/kcp-dev/kcp/pkg/syncer/shared"
	"github.com/tektoncd/results/pkg/server/api/v1alpha2/result"
	"google.golang.org/grpc/metadata"
	"time"

	"github.com/tektoncd/results/pkg/watcher/convert"
	"github.com/tektoncd/results/pkg/watcher/reconciler"
	"github.com/tektoncd/results/pkg/watcher/reconciler/annotation"
	"github.com/tektoncd/results/pkg/watcher/results"
	rpb "github.com/tektoncd/results/proto/results/v1alpha2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/logging"
)

var (
	clock = clockwork.NewRealClock()
)

// Reconciler implements common reconciler behavior across different Tekton Run
// Object types.
type Reconciler struct {
	resultsClient *results.Client
	objectClient  ObjectClient
	cfg           *reconciler.Config
	enqueue       func(interface{}, time.Duration)
}

// NewDynamicReconciler creates a new dynamic Reconciler.
func NewDynamicReconciler(rc rpb.ResultsClient, oc ObjectClient, cfg *reconciler.Config, enqueue func(interface{}, time.Duration)) *Reconciler {
	return &Reconciler{
		resultsClient: &results.Client{ResultsClient: rc},
		objectClient:  oc,
		cfg:           cfg,
		enqueue:       enqueue,
	}
}

// Reconcile handles result/record uploading for the given Run object.
// If enabled, the object may be deleted upon successful result upload.
func (r *Reconciler) Reconcile(ctx context.Context, object results.Object) error {
	log := logging.FromContext(ctx)

	if object.GetObjectKind().GroupVersionKind().Empty() {
		gvk, err := convert.InferGVK(object)
		if err != nil {
			return err
		}
		object.GetObjectKind().SetGroupVersionKind(gvk)
		log.Infof("Post-GVK Object: %v", object)
	}

	// Get namespace name from object
	ns, err := r.cfg.KubeClient.CoreV1().Namespaces().Get(ctx, object.GetNamespace(), metav1.GetOptions{})
	if err != nil {
		log.Errorf("error fetching namespace: %v", err)
		return err
	}

	// Get KCP namespace locator
	nl, ok, err := shared.LocatorFromAnnotations(ns.Annotations)
	if err != nil {
		log.Errorf("error getting cluster and namespace: %v", err)
		return err
	}
	if !ok {
		log.Info("skipping resource: object not in KCP namespace")
		return nil
	}

	// Add service account token to metadata when authMode is "service-account"
	if r.cfg.Auth.AuthMode == "service-account" {
		sanl := *nl
		sanl.Namespace = r.cfg.Auth.ServiceAccountNamespace
		sans, err := shared.PhysicalClusterNamespaceName(sanl)
		if err != nil {
			log.Errorf("cannot determine KCP service account workspace: %v", err)
			return err
		}

		sa, err := r.cfg.KubeClient.CoreV1().ServiceAccounts(sans).Get(ctx, r.cfg.Auth.ServiceAccount, metav1.GetOptions{})
		if err != nil {
			log.Errorf("cannot find service account in KCP workspace %s : %v", sans, err)
			return err
		}
		if len(sa.Secrets) == 0 {
			log.Errorf("cannot find secret in service account config: %v", err)
			return err
		}

		s, err := r.cfg.KubeClient.CoreV1().Secrets(sans).Get(ctx, sa.Secrets[0].Name, metav1.GetOptions{})
		if err != nil {
			log.Errorf("cannot find service account secret in KCP workspace %s : %v", sans, err)
			return err
		}

		t, ok := s.Data["token"]
		if !ok {
			log.Errorf("cannot find token in service account secret config: %v", err)
			return err
		}

		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+string(t))
	}

	// Format parent as per API specs
	parent := result.FormatParent(nl.Workspace.String(), nl.Namespace)

	// Update record.
	res, rec, err := r.resultsClient.Put(ctx, parent, object)
	if err != nil {
		log.Errorf("error updating Record: %v", err)
		return err
	}

	// Update object with Result Annotations.
	if a := object.GetAnnotations(); !r.cfg.GetDisableAnnotationUpdate() {
		if res.GetName() != a[annotation.Result] || rec.GetName() != a[annotation.Record] {
			patch, err := annotation.Add(res.GetName(), rec.GetName())
			if err != nil {
				log.Errorf("error adding Result annotations: %v", err)
				return err
			}
			if err := r.objectClient.Patch(ctx, object.GetName(), types.MergePatchType, patch, metav1.PatchOptions{}); err != nil {
				log.Errorf("Patch: %v", err)
				return err
			}
		} else {
			log.Info("skipping CRD patch: annotations already match")
		}
	}

	// If the Object is complete and not yet marked for deletion, cleanup the run resource from the cluster.
	done := isDone(object)
	log.Infof("should skipping resource deletion?  - done: %t, delete enabled: %t", done, r.cfg.GetCompletedResourceGracePeriod() != 0)
	if done && r.cfg.GetCompletedResourceGracePeriod() != 0 {
		if o := object.GetOwnerReferences(); len(o) > 0 {
			log.Infof("resource is owned by another object, defering deletion to parent resource(s): %v", o)
			return nil
		}

		// We haven't hit the grace period yet - reenqueue the key for processing later.
		if s := clock.Since(rec.GetUpdateTime().AsTime()); s < r.cfg.GetCompletedResourceGracePeriod() {
			log.Infof("object is not ready for deletion - grace period: %v, time since completion: %v", r.cfg.GetCompletedResourceGracePeriod(), s)
			r.enqueue(object, r.cfg.GetCompletedResourceGracePeriod())
			return nil
		}
		log.Infof("deleting PipelineRun UID %s", object.GetUID())
		if err := r.objectClient.Delete(ctx, object.GetName(), metav1.DeleteOptions{
			Preconditions: metav1.NewUIDPreconditions(string(object.GetUID())),
		}); err != nil && !errors.IsNotFound(err) {
			log.Errorf("PipelineRun.Delete: %v", err)
			return err
		}
	} else {
		log.Infof("skipping resource deletion - done: %t, delete enabled: %t, %v", done, r.cfg.GetCompletedResourceGracePeriod() != 0, r.cfg.GetCompletedResourceGracePeriod())
	}
	return nil
}

func isDone(o results.Object) bool {
	return o.GetStatusCondition().GetCondition(apis.ConditionSucceeded).IsTrue()
}
