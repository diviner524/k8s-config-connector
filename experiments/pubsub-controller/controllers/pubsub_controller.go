/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/cbroglie/mustache"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1alpha1 "pubsub-controller/api/v1alpha1"
)

// PubSubReconciler reconciles a PubSub object
type PubSubReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	OrderedResources []*ManagedResource
}

// We should probably add a Ready field to indicate how we can tell if a KRM resource is ready
type Resource struct {
	Name       string   `yaml:"name"`
	Template   string   `yaml:"template"`
	Conditions string   `yaml:"conditions,omitempty"`
	DependsOn  []string `yaml:"dependsOn,omitempty"`
}

type ResourcesWrapper struct {
	Resources []Resource `json:"resources"`
}

// ManagedResource is a struct which has several fields to describe a Resource being managed as a node in the DAG
// Resource is a reference to the actual Resource Struct
// Dependencies is a list of references to other Resources which this Resource depends on
// IsReady is a boolean which indicates whether this Resource is ready or not
// TODO: It is the responsibility of the Bootstrap controller (Not this controller) to build the Dependencies list which includes both explict and implicit dependencies
type ManagedResource struct {
	Resource     *Resource
	Dependencies []*ManagedResource
	IsReady      bool
}

// +kubebuilder:rbac:groups=infra.sample.org,resources=pubsubs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infra.sample.org,resources=pubsubs/status,verbs=get;update;patch

func (r *PubSubReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("pubsub", req.NamespacedName)

	u := &unstructured.Unstructured{}
	u.SetAPIVersion("infra.sample.org/v1alpha1")
	u.SetKind("PubSub")

	if err := r.Get(ctx, req.NamespacedName, u); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("resource not found in API server; finishing reconcile", "resource", req.NamespacedName)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	values := make(map[string]interface{})
	values["root"] = u.UnstructuredContent()

	// loop through r.OrderedResources and create each resource
	for _, resource := range r.OrderedResources {
		r.Log.Info("Creating resource", "name", resource.Resource.Name)
		// Check if the depenencies of the resource are UpToDate k8s resources
		for _, dependency := range resource.Dependencies {
			if !dependency.IsReady {
				r.Log.Info("Dependency not ready", "dependency", dependency.Resource.Name)
				return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
			}
		}

		// Render of the template. Topological ordering ensures that all the values from dependencies are added to the values map
		renderedTemplate, _ := mustache.Render(resource.Resource.Template, values)
		r.Log.Info("Resource template", "template", renderedTemplate)

		// Convert renderedTemplate to unstructured object and create the resource through r.Create()
		decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

		obj := &unstructured.Unstructured{}
		_, _, err := decUnstructured.Decode([]byte(renderedTemplate), nil, obj)
		if err != nil {
			fmt.Printf("Error decoding YAML: %v", err)
			return reconcile.Result{}, err
		}
		if err := copyNamespace(u, obj); err != nil {
			return reconcile.Result{}, err
		}
		if err := r.Create(ctx, obj); err != nil && !apierrors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		}

		nn := types.NamespacedName{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
		}

		ready, err := r.isObjectReady(ctx, nn, obj)
		if err != nil {
			return reconcile.Result{}, err
		}

		// TODO: This is a hack. We should store the status of managed resource in the KRM resource (from req), instead of persisting it in the reconciler.
		resource.IsReady = ready
		r.Log.Info("Resource readiness", "name", resource.Resource.Name, "ready", resource.IsReady)

		// If the resource is ready, add resource as a map into the values map
		if ready {
			resourceMap := obj.UnstructuredContent()
			values[resource.Resource.Name] = resourceMap
		}

	}

	return ctrl.Result{}, nil
}

func (r *PubSubReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1alpha1.PubSub{}).
		Complete(r)
}

func copyNamespace(u, obj *unstructured.Unstructured) error {
	namespace, found, err := unstructured.NestedString(u.Object, "metadata", "namespace")
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("namespace not found in the source object")
	}

	if err := unstructured.SetNestedField(obj.Object, namespace, "metadata", "namespace"); err != nil {
		return err
	}

	return nil
}

func (r *PubSubReconciler) isObjectReady(ctx context.Context, nn types.NamespacedName, u *unstructured.Unstructured) (bool, error) {
	if err := r.Get(ctx, nn, u); err != nil {
		if errors.IsNotFound(err) {
			return false, nil // Object is not found, hence not ready
		}
		return false, err // An error occurred
	}

	conditions, found, err := unstructured.NestedSlice(u.Object, "status", "conditions")
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil // "conditions" not found, hence not ready
	}

	for _, cond := range conditions {
		condition, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		if condition["type"] == "Ready" && condition["status"] == "True" {
			return true, nil // Object is ready
		}
	}

	return false, nil // "Ready" condition not met
}
