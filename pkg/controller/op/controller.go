// Copyright 2022 Google LLC
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

package tf

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/controller/jitter"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/controller/lifecyclehandler"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/controller/metrics"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/controller/predicate"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/controller/ratelimiter"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/controller/resourceactuation"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/execution"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/k8s"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/krmtotf"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/lease/leaser"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/runtime"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/servicemapping/servicemappingloader"

	"github.com/go-logr/logr"
	tfschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/sync/semaphore"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	klog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var logger = klog.Log

type Reconciler struct {
	lifecyclehandler.LifecycleHandler
	metrics.ReconcilerMetrics
	resourceLeaser *leaser.ResourceLeaser
	mgr            manager.Manager
	schemaRef      *k8s.SchemaReference
	schemaRefMu    sync.RWMutex
	provider       *tfschema.Provider
	smLoader       *servicemappingloader.ServiceMappingLoader
	logger         logr.Logger
	// Fields used for triggering reconciliations when dependencies are ready
	immediateReconcileRequests chan event.GenericEvent
	resourceWatcherRoutines    *semaphore.Weighted // Used to cap number of goroutines watching unready dependencies
}

func Add(mgr manager.Manager, crd *apiextensions.CustomResourceDefinition, provider *tfschema.Provider, smLoader *servicemappingloader.ServiceMappingLoader) (k8s.SchemaReferenceUpdater, error) {
	kind := crd.Spec.Names.Kind
	apiVersion := k8s.GetAPIVersionFromCRD(crd)
	controllerName := fmt.Sprintf("%v-controller", strings.ToLower(kind))
	immediateReconcileRequests := make(chan event.GenericEvent, k8s.ImmediateReconcileRequestsBufferSize)
	resourceWatcherRoutines := semaphore.NewWeighted(k8s.MaxNumResourceWatcherRoutines)
	r, err := NewReconciler(mgr, crd, provider, smLoader, immediateReconcileRequests, resourceWatcherRoutines)
	if err != nil {
		return nil, err
	}
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       kind,
			"apiVersion": apiVersion,
		},
	}
	_, err = builder.
		ControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{MaxConcurrentReconciles: k8s.ControllerMaxConcurrentReconciles, RateLimiter: ratelimiter.NewRateLimiter()}).
		Watches(&source.Channel{Source: immediateReconcileRequests}, &handler.EnqueueRequestForObject{}).
		For(obj, builder.OnlyMetadata, builder.WithPredicates(predicate.UnderlyingResourceOutOfSyncPredicate{})).
		Build(r)
	if err != nil {
		return nil, fmt.Errorf("error creating new controller: %v", err)
	}
	logger.Info("Registered controller", "kind", kind, "apiVersion", apiVersion)
	return r, nil
}

func NewReconciler(mgr manager.Manager, crd *apiextensions.CustomResourceDefinition, p *tfschema.Provider, smLoader *servicemappingloader.ServiceMappingLoader, immediateReconcileRequests chan event.GenericEvent, resourceWatcherRoutines *semaphore.Weighted) (*Reconciler, error) {
	controllerName := fmt.Sprintf("%v-controller", strings.ToLower(crd.Spec.Names.Kind))
	return &Reconciler{
		LifecycleHandler: lifecyclehandler.NewLifecycleHandler(
			mgr.GetClient(),
			mgr.GetEventRecorderFor(controllerName),
		),
		resourceLeaser: leaser.NewResourceLeaser(p, smLoader, mgr.GetClient()),
		mgr:            mgr,
		schemaRef: &k8s.SchemaReference{
			CRD:        crd,
			JsonSchema: k8s.GetOpenAPIV3SchemaFromCRD(crd),
			GVK: schema.GroupVersionKind{
				Group:   crd.Spec.Group,
				Version: k8s.GetVersionFromCRD(crd),
				Kind:    crd.Spec.Names.Kind,
			},
		},
		ReconcilerMetrics: metrics.ReconcilerMetrics{
			ResourceNameLabel: metrics.ResourceNameLabel,
		},
		provider:                   p,
		smLoader:                   smLoader,
		logger:                     logger.WithName(controllerName),
		immediateReconcileRequests: immediateReconcileRequests,
		resourceWatcherRoutines:    resourceWatcherRoutines,
	}, nil
}

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (res reconcile.Result, err error) {
	r.schemaRefMu.RLock()
	defer r.schemaRefMu.RUnlock()
	r.logger.Info("starting reconcile", "resource", req.NamespacedName)
	startTime := time.Now()
	r.RecordReconcileWorkers(ctx, r.schemaRef.GVK)
	defer r.AfterReconcile()
	defer r.RecordReconcileMetrics(ctx, r.schemaRef.GVK, req.Namespace, req.Name, startTime, &err)

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(r.schemaRef.GVK)

	if err := r.Get(ctx, req.NamespacedName, u); err != nil {
		if apierrors.IsNotFound(err) {
			r.logger.Info("resource not found in API server; finishing reconcile", "resource", req.NamespacedName)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	skip, err := resourceactuation.ShouldSkip(u)
	if err != nil {
		return reconcile.Result{}, err
	}
	if skip {
		r.logger.Info("Skipping reconcile as nothing has changed and 0 reconcile period is set", "resource", req.NamespacedName)
		return reconcile.Result{}, nil
	}
	sm, err := r.smLoader.GetServiceMapping(u.GroupVersionKind().Group)
	if err != nil {
		return reconcile.Result{}, err
	}
	u, err = k8s.TriggerManagedFieldsMetadata(ctx, r.Client, u)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("error triggering Server-Side Apply (SSA) metadata: %w", err)
	}
	cvt := runtime.NewConverter(u.GetKind())
	resource, err := krmtotf.NewResource(u, sm, r.provider)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not parse resource %s: %v", req.NamespacedName.String(), err)
	}

	requeue, err := r.sync(ctx, u, cvt)
	if err != nil {
		return reconcile.Result{}, err
	}
	if requeue {
		return reconcile.Result{Requeue: true}, nil
	}
	jitteredPeriod, err := jitter.GenerateJitteredReenqueuePeriod(r.schemaRef.GVK, r.smLoader, nil, u)
	if err != nil {
		return reconcile.Result{}, err
	}
	r.logger.Info("successfully finished reconcile", "resource", k8s.GetNamespacedName(resource), "time to next reconciliation", jitteredPeriod)
	return reconcile.Result{RequeueAfter: jitteredPeriod}, nil
}

func (r *Reconciler) sync(ctx context.Context, u *unstructured.Unstructured, cvt runtime.OnePlatformConverter) (requeue bool, err error) {
	// isolate any panics to only this function
	defer execution.RecoverWithInternalError(&err)
	krmResource, _ := k8s.NewResource(u)
	if !u.GetDeletionTimestamp().IsZero() {
		// Deleting
		r.logger.Info("finalizing resource deletion", "resource", k8s.GetNamespacedName(u))
		if !k8s.HasFinalizer(u, k8s.ControllerFinalizerName) {
			r.logger.Info("no controller finalizer is present; no finalization necessary",
				"resource", k8s.GetNamespacedName(u))
			return false, nil
		}
		if k8s.HasFinalizer(u, k8s.DeletionDefenderFinalizerName) {
			r.logger.Info("deletion defender has not yet been finalized; requeuing", "resource", k8s.GetNamespacedName(u))
			return true, nil
		}
		if err := r.HandleDeleting(ctx, krmResource); err != nil {
			return false, err
		}
		if k8s.HasAbandonAnnotation(krmResource) {
			r.logger.Info("deletion policy set to abandon; abandoning underlying resource", "resource", k8s.GetNamespacedName(krmResource))
			return false, r.HandleDeleted(ctx, krmResource)

		}
		remoteObj, err := cvt.GetResource(ctx, u)
		if err != nil {
			return false, err
		}
		if remoteObj == nil {
			r.logger.Info("underlying resource does not exist; no API call necessary", "resource", k8s.GetNamespacedName(krmResource))
			return false, r.HandleDeleted(ctx, krmResource)
		}
		r.logger.Info("deleting underlying resource", "resource", k8s.GetNamespacedName(krmResource))
		err = cvt.DeleteResource(ctx, u)
		if err != nil {
			return false, err
		}
		return false, r.HandleDeleted(ctx, krmResource)
	}
	remoteObj, _ := cvt.GetResource(ctx, u)
	// ensure the finalizers before apply
	if err := r.EnsureFinalizers(ctx, krmResource, krmResource, k8s.ControllerFinalizerName, k8s.DeletionDefenderFinalizerName); err != nil {
		return false, err
	}
	if remoteObj == nil {
		r.logger.Info("creating/updating underlying resource", "resource", k8s.GetNamespacedName(krmResource))
		if err := r.HandleUpdating(ctx, krmResource); err != nil {
			return false, err
		}
		// Create the resource through converter cvt
		remoteObj, err = cvt.CreateResource(ctx, u)
		krmResource, _ = k8s.NewResource(remoteObj)
	} else {
		// Get diff through cvt
		diff, err := cvt.GetDiff(ctx, u, remoteObj)
		if err != nil {
			return false, err
		}
		if diff {
			r.logger.Info("creating/updating underlying resource", "resource", k8s.GetNamespacedName(krmResource))
			if err := r.HandleUpdating(ctx, krmResource); err != nil {
				return false, err
			}
			// Update the resource through converter cvt
			remoteObj, err = cvt.UpdateResource(ctx, u)
			krmResource, _ = k8s.NewResource(remoteObj)
			if err != nil {
				return false, err
			}
		} else {
			r.logger.Info("underlying resource already up to date", "resource", k8s.GetNamespacedName(krmResource))
		}
	}
	return false, r.HandleUpToDate(ctx, krmResource)
}

var _ k8s.SchemaReferenceUpdater = &Reconciler{}

func (r *Reconciler) UpdateSchema(crd *apiextensions.CustomResourceDefinition) error {
	r.schemaRefMu.Lock()
	defer r.schemaRefMu.Unlock()
	return k8s.UpdateSchema(r.schemaRef, crd)
}
